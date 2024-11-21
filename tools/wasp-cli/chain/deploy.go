// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"context"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/clients"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet/wallets"

	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/origin"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"
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

func initDeployMoveContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy-move-contract",
		Short: "Deploy a new move contract and save its package id",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
			defer cancel()

			l1Client := cliclients.L1Client()
			kp := wallet.Load()
			packageID, err := l1Client.DeployISCContracts(ctx, cryptolib.SignerToIotaSigner(kp))
			log.Check(err)

			config.SetPackageID(packageID)

			log.Printf("Move contract deployed.\nPackageID: %v\n", packageID.String())
		},
	}

	return cmd
}

func initializeNewChainState(stateController *cryptolib.Address, gasCoinObjectID iotago.ObjectID) *transaction.StateMetadata {
	initParams := origin.DefaultInitParams(isc.NewAddressAgentID(stateController)).Encode()
	store := indexedstore.New(state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB()))
	// TODO: Place GasCoinObjectID into here once VMs part is done
	_, stateMetadata := origin.InitChain(allmigrations.LatestSchemaVersion, store, initParams, 0, isc.BaseTokenCoinInfo)
	return stateMetadata
}

func createGasCoin(ctx context.Context, client clients.L1Client, wallet wallets.Wallet) (iotago.ObjectID, error) {
	coins, err := client.GetCoinObjsForTargetAmount(ctx, wallet.Address().AsIotaAddress(), 10*isc.Million)
	if err != nil {
		return iotago.ObjectID{}, err
	}

	if len(coins) == 0 {
		return iotago.ObjectID{}, errors.New("no coins found")
	}

	splitCoins, err := client.SplitCoin(ctx, iotaclient.SplitCoinRequest{
		Coin:      coins[0].CoinObjectID,
		Signer:    wallet.Address().AsIotaAddress(),
		GasBudget: iotajsonrpc.NewBigInt(10 * isc.Million),
		Gas:       coins[1].CoinObjectID,
		SplitAmounts: []*iotajsonrpc.BigInt{
			iotajsonrpc.NewBigInt(1 * isc.Million),
		},
	})
	if err != nil {
		return iotago.ObjectID{}, err
	}

	result, err := client.SignAndExecuteTransaction(ctx, cryptolib.SignerToIotaSigner(wallet), splitCoins.TxBytes, &iotajsonrpc.IotaTransactionBlockResponseOptions{
		ShowEffects:        true,
		ShowBalanceChanges: true,
	})
	if err != nil {
		return iotago.ObjectID{}, err
	}

	if len(result.Effects.Data.V1.Created) != 1 {
		return iotago.ObjectID{}, errors.New("mom, help")
	}

	return *result.Effects.Data.V1.Created[0].Reference.ObjectID, nil
}

func transferGasCoinToCommittee(ctx context.Context, client clients.L1Client, wallet wallets.Wallet, coinObjectID iotago.ObjectID, committeeAddress *cryptolib.Address) error {
	transferTx, err := client.TransferObject(ctx, iotaclient.TransferObjectRequest{
		ObjectID:  &coinObjectID,
		Recipient: committeeAddress.AsIotaAddress(),
		GasBudget: iotajsonrpc.NewBigInt(10 * isc.Million),
		Signer:    wallet.Address().AsIotaAddress(),
	})

	if err != nil {
		return err
	}

	_, err = client.SignAndExecuteTransaction(ctx, cryptolib.SignerToIotaSigner(wallet), transferTx.TxBytes, &iotajsonrpc.IotaTransactionBlockResponseOptions{})
	if err != nil {
		return err
	}

	return nil
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

			gasCoin, err := createGasCoin(ctx, l1Client, kp)
			log.Check(err)

			// TODO: We need to decide if we want to deploy a new contract for each new chain, or use one constant for it.
			//packageID, err := l1Client.DeployISCContracts(ctx, cryptolib.SignerToIotaSigner(kp))
			packageID := config.GetPackageID()

			stateController := doDKG(ctx, node, peers, quorum)

			stateMetadata := initializeNewChainState(stateController, gasCoin)

			par := apilib.CreateChainParams{
				Layer1Client:      l1Client,
				CommitteeAPIHosts: config.NodeAPIURLs([]string{node}),
				N:                 uint16(len(node)),
				T:                 uint16(quorum),
				OriginatorKeyPair: kp,
				Textout:           os.Stdout,
				PackageID:         packageID,
				StateMetadata:     *stateMetadata,
			}

			// Transfer Gas Coin to Committee
			err = transferGasCoinToCommittee(ctx, l1Client, kp, gasCoin, stateController)
			log.Check(err)

			// Deploy final chain
			chainID, err := apilib.DeployChain(ctx, par, stateController)
			log.Check(err)

			config.AddChain(chainName, chainID.String())

			activateChain(node, chainName, chainID)
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
