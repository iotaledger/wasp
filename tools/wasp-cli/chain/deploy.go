// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/iotaledger/wasp/clients"
	"github.com/iotaledger/wasp/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/components/app"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/origin"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

func GetAllWaspNodes() []int {
	ret := []int{}
	for index := range viper.GetStringMap("wasp") {
		i, err := strconv.Atoi(index)
		log.Check(err)
		ret = append(ret, i)
	}
	return ret
}

func controllerAddrDefaultFallback(addr string) *cryptolib.Address {
	if addr == "" {
		return wallet.Load().Address()
	}
	govControllerAddr, err := cryptolib.NewAddressFromHexString(addr)
	log.Check(err)
	panic("refactor me: what are we doing without network prefixes here?")
	/*if parameters.Bech32Hrp != parameters.NetworkPrefix(prefix) {
		log.Fatalf("unexpected prefix. expected: %s, actual: %s", parameters.Bech32Hrp, prefix)
	}*/
	return govControllerAddr
}

func deployISCMoveContract(ctx context.Context, client clients.L1Client, signer cryptolib.Signer) (iotago.PackageID, error) {
	iscBytecode := contracts.ISC()

	txnBytes, err := client.Publish(ctx, iotaclient.PublishRequest{
		Sender:          signer.Address().AsSuiAddress(),
		CompiledModules: iscBytecode.Modules,
		Dependencies:    iscBytecode.Dependencies,
		GasBudget:       iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget * 10),
	})

	if err != nil {
		return iotago.PackageID{}, err
	}

	txnResponse, err := client.SignAndExecuteTransaction(
		ctx,
		cryptolib.SignerToSuiSigner(signer),
		txnBytes.TxBytes,
		&iotajsonrpc.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	if err != nil {
		return iotago.PackageID{}, err
	}

	packageId, err := txnResponse.GetPublishedPackageID()

	if err != nil {
		return iotago.PackageID{}, err
	}

	if packageId == nil {
		return iotago.PackageID{}, errors.Errorf("no published package ID in response")
	}

	return *packageId, err
}

func initDeployCmd() *cobra.Command {
	var (
		node             string
		peers            []string
		quorum           int
		evmChainID       uint16
		blockKeepAmount  int32
		govControllerStr string
		chainName        string
	)

	cmd := &cobra.Command{
		Use:   "deploy --chain=<name>",
		Short: "Deploy a new chain",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			chainName = defaultChainFallback(chainName)
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
			defer cancel()

			if !util.IsSlug(chainName) {
				log.Fatalf("invalid chain name: %s, must be in slug format, only lowercase and hyphens, example: foo-bar", chainName)
			}

			l1Client := cliclients.L1Client()
			kp := wallet.Load()

			packageID, err := deployISCMoveContract(ctx, l1Client, kp)
			log.Check(err)

			govController := controllerAddrDefaultFallback(govControllerStr)

			// TODO: Fixme: doDKG requires a somewhat runnable wasp node :D
			// stateController := doDKG(ctx, node, peers, quorum)

			stateController := cryptolib.NewRandomAddress()

			par := apilib.CreateChainParams{
				Layer1Client:         l1Client,
				CommitteeAPIHosts:    config.NodeAPIURLs([]string{node}),
				N:                    uint16(len(node)),
				T:                    uint16(quorum),
				OriginatorKeyPair:    wallet.Load(),
				Textout:              os.Stdout,
				GovernanceController: govController,
				PackageID:            packageID,
				InitParams: dict.Dict{
					origin.ParamChainOwner:      isc.NewAddressAgentID(govController).Bytes(),
					origin.ParamEVMChainID:      codec.Encode[uint16](evmChainID),
					origin.ParamBlockKeepAmount: codec.Encode[int32](blockKeepAmount),
					origin.ParamWaspVersion:     codec.Encode[string](app.Version),
				},
			}

			chainID, err := apilib.DeployChain(ctx, par, stateController, govController)
			log.Check(err)

			config.AddChain(chainName, chainID.String())

			// TODO: Fixme: This requires a runnable node as well.
			// activateChain(node, chainName, chainID)
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)
	waspcmd.WithPeersFlag(cmd, &peers)
	cmd.Flags().Uint16VarP(&evmChainID, "evm-chainid", "", evm.DefaultChainID, "ChainID")
	cmd.Flags().Int32VarP(&blockKeepAmount, "block-keep-amount", "", governance.DefaultBlockKeepAmount, "Amount of blocks to keep in the blocklog (-1 to keep all blocks)")
	cmd.Flags().StringVar(&chainName, "chain", "", "name of the chain")
	log.Check(cmd.MarkFlagRequired("chain"))
	cmd.Flags().IntVar(&quorum, "quorum", 0, "quorum (default: 3/4s of the number of committee nodes)")
	cmd.Flags().StringVar(&govControllerStr, "gov-controller", "", "governance controller address")
	return cmd
}
