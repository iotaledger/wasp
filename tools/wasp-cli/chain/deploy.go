// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/packages/apilib"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/kvstore/mapdb"
	"github.com/iotaledger/wasp/v2/packages/origin"
	"github.com/iotaledger/wasp/v2/packages/parameters"
	"github.com/iotaledger/wasp/v2/packages/state/indexedstore"
	"github.com/iotaledger/wasp/v2/packages/state/statetest"
	"github.com/iotaledger/wasp/v2/packages/transaction"
	"github.com/iotaledger/wasp/v2/packages/util"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
	"github.com/iotaledger/wasp/v2/packages/vm/core/migrations/allmigrations"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/wallet"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/wallet/wallets"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/waspcmd"
)

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

			log.Printf("Move contract deployed.\nPackageID: %v\n", packageID.String())
		},
	}

	return cmd
}

func initializeNewChainState(chainAdmin *cryptolib.Address, gasCoinObject iotago.ObjectID, l1Params *parameters.L1Params) *transaction.StateMetadata {
	initParams := origin.DefaultInitParams(isc.NewAddressAgentID(chainAdmin)).Encode()
	store := indexedstore.New(statetest.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB()))
	_, stateMetadata := origin.InitChain(allmigrations.LatestSchemaVersion, store, initParams, gasCoinObject, isc.GasCoinTargetValue, l1Params)
	return stateMetadata
}

func CreateAndSendGasCoin(ctx context.Context, client clients.L1Client, wallet wallets.Wallet, committeeAddress *iotago.Address, l1Params *parameters.L1Params) (iotago.ObjectID, error) {
	coins, err := client.GetCoinObjsForTargetAmount(ctx, wallet.Address().AsIotaAddress(), isc.GasCoinTargetValue, isc.GasCoinTargetValue)
	if err != nil {
		return iotago.ObjectID{}, fmt.Errorf("GasCoin with targeting blanace not found: %w", err)
	}

	txb := iotago.NewProgrammableTransactionBuilder()
	splitCoinCmd := txb.Command(
		iotago.Command{
			SplitCoins: &iotago.ProgrammableSplitCoins{
				Coin:    iotago.GetArgumentGasCoin(),
				Amounts: []iotago.Argument{txb.MustPure(isc.GasCoinTargetValue)},
			},
		},
	)

	txb.TransferArg(committeeAddress, splitCoinCmd)

	txData := iotago.NewProgrammable(
		wallet.Address().AsIotaAddress(),
		txb.Finish(),
		[]*iotago.ObjectRef{coins[0].Ref()},
		uint64(isc.GasCoinTargetValue),
		l1Params.Protocol.ReferenceGasPrice.Uint64(),
	)

	txnBytes, err := bcs.Marshal(&txData)
	if err != nil {
		return iotago.ObjectID{}, err
	}

	result, err := client.SignAndExecuteTransaction(
		ctx,
		&iotaclient.SignAndExecuteTransactionRequest{
			Signer:      cryptolib.SignerToIotaSigner(wallet),
			TxDataBytes: txnBytes,
			Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
				ShowEffects:       true,
				ShowObjectChanges: true,
			},
		},
	)
	if err != nil {
		return iotago.ObjectID{}, fmt.Errorf("failed to create GasCoin: %w", err)
	}

	gasCoin, err := result.GetCreatedCoinByType("iota", "IOTA")
	if err != nil {
		return iotago.ObjectID{}, err
	}

	return *gasCoin.ObjectID, nil
}

type chainInitResult struct {
	gasCoinObject    iotago.ObjectID
	l1Params         *parameters.L1Params
	committeeAddress *cryptolib.Address
}

func initializeDeploymentWithGasCoin(ctx context.Context, signer wallets.Wallet, node string, chainName string, peers []string, quorum int) (*chainInitResult, error) {
	if !util.IsSlug(chainName) {
		log.Fatalf("invalid chain name: %s, must be in slug format, only lowercase and hyphens, example: foo-bar", chainName)
	}

	l1Client := cliclients.L1Client()

	client := cliclients.WaspClientWithVersionCheck(ctx, node)
	_, header, err := client.ChainsAPI.GetChainInfo(ctx).Execute()
	defer header.Body.Close()

	// We expect a 404 if no chain has been deployed yet. In any other case, show the error.
	if err != nil && !strings.Contains(err.Error(), strconv.Itoa(http.StatusNotFound)) {
		return nil, fmt.Errorf("failed to get current chain info: %w", err)
	}

	// Now check if the response is 404, if not, a Chain has already been deployed. Exit early.
	if header != nil && header.StatusCode != http.StatusNotFound {
		return nil, fmt.Errorf("a chain has already been deployed")
	}

	committeeAddr := doDKG(ctx, node, peers, quorum)

	l1Params, err := parameters.FetchLatest(ctx, l1Client.IotaClient())
	log.Check(err)

	gasCoin, err := CreateAndSendGasCoin(ctx, l1Client, signer, committeeAddr.AsIotaAddress(), l1Params)
	log.Check(err)

	return &chainInitResult{
		l1Params:         l1Params,
		gasCoinObject:    gasCoin,
		committeeAddress: committeeAddr,
	}, nil
}

func finalizeChainDeployment(ctx context.Context, node string, packageID *iotago.PackageID, chainInitResult chainInitResult, stateMetadata *transaction.StateMetadata) isc.ChainID {
	par := apilib.CreateChainParams{
		Layer1Client:      cliclients.L1Client(),
		CommitteeAPIHosts: config.NodeAPIURLs([]string{node}),
		Signer:            wallet.Load(),
		Textout:           os.Stdout,
		PackageID:         *packageID,
		StateMetadata:     *stateMetadata,
	}

	chainID, err := apilib.DeployChain(ctx, par, chainInitResult.committeeAddress)
	log.Check(err)

	return chainID
}

func initDeployCmd() *cobra.Command {
	var (
		node             string
		peers            []string
		quorum           int
		iscPackageIDStr  string
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
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
			defer cancel()

			kp := wallet.Load()

			var iscPackageID = &iotago.PackageID{}
			var err error

			if iscPackageIDStr != "" {
				iscPackageID, err = iotago.PackageIDFromHex(iscPackageIDStr)
				log.Check(err)
			} else {
				log.Printf("Deploying Move contract...\n")
				l1Client := cliclients.L1Client()
				*iscPackageID, err = l1Client.DeployISCContracts(ctx, cryptolib.SignerToIotaSigner(kp))
				log.Check(err)
			}

			node = waspcmd.DefaultWaspNodeFallback(node)
			chainName = defaultChainFallback(chainName)

			result, err := initializeDeploymentWithGasCoin(ctx, kp, node, chainName, peers, quorum)
			if err != nil {
				log.Fatal(err)
			}
			stateMetadata := initializeNewChainState(kp.Address(), result.gasCoinObject, result.l1Params)
			chainID := finalizeChainDeployment(ctx, node, iscPackageID, *result, stateMetadata)

			config.AddChain(chainName, chainID.String())
			activateChain(ctx, node, chainName, chainID)
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)
	waspcmd.WithPeersFlag(cmd, &peers)
	cmd.Flags().StringVar(&iscPackageIDStr, "package-id", "", "ISC L1 package ID. If not set, new package will be deployed (see `deploy-move-contract` command)")
	cmd.Flags().Uint16VarP(&evmChainID, "evm-chainid", "", evm.DefaultChainID, "ChainID")
	cmd.Flags().Int32VarP(&blockKeepAmount, "block-keep-amount", "", governance.DefaultBlockKeepAmount, "Amount of blocks to keep in the blocklog (-1 to keep all blocks)")
	cmd.Flags().StringVar(&chainName, "chain", "", "name of the chain")
	log.Check(cmd.MarkFlagRequired("chain"))
	cmd.Flags().IntVar(&quorum, "quorum", 0, "quorum (default: 3/4s of the number of committee nodes)")
	cmd.Flags().StringVar(&govControllerStr, "gov-controller", "", "governance controller address")
	return cmd
}
