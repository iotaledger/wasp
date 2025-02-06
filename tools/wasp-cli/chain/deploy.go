// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"context"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/iotaledger/hive.go/kvstore/mapdb"

	"github.com/iotaledger/wasp/clients"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/origin"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"

	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet/wallets"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
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

			config.SetPackageID(packageID)

			log.Printf("Move contract deployed.\nPackageID: %v\n", packageID.String())
		},
	}

	return cmd
}

func initializeNewChainState(stateController *cryptolib.Address, gasCoinObject iotago.ObjectID) *transaction.StateMetadata {
	initParams := origin.DefaultInitParams(isc.NewAddressAgentID(stateController)).Encode()
	store := indexedstore.New(state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB()))
	_, stateMetadata := origin.InitChain(allmigrations.LatestSchemaVersion, store, initParams, gasCoinObject, isc.GasCoinTargetValue, isc.BaseTokenCoinInfo)
	return stateMetadata
}

func createAndSendGasCoin(ctx context.Context, client clients.L1Client, wallet wallets.Wallet, committeeAddress *iotago.Address) (iotago.ObjectID, error) {

	coins, err := client.GetCoinObjsForTargetAmount(ctx, wallet.Address().AsIotaAddress(), isc.GasCoinTargetValue, isc.GasCoinTargetValue)
	if err != nil {
		return iotago.ObjectID{}, err
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
		parameters.L1().Protocol.ReferenceGasPrice.Uint64(),
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
		return iotago.ObjectID{}, err
	}

	gasCoin, err := result.GetCreatedCoin("iota", "IOTA")
	if err != nil {
		return iotago.ObjectID{}, err
	}

	return *gasCoin.ObjectID, nil
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

			// TODO: We need to decide if we want to deploy a new contract for each new chain, or use one constant for it.
			// packageID, err := l1Client.DeployISCContracts(ctx, cryptolib.SignerToIotaSigner(kp))
			packageID := config.GetPackageID()

			stateControllerAddress := doDKG(ctx, node, peers, quorum)

			gasCoin, err := createAndSendGasCoin(ctx, l1Client, kp, stateControllerAddress.AsIotaAddress())
			log.Check(err)

			stateMetadata := initializeNewChainState(stateControllerAddress, gasCoin)

			par := apilib.CreateChainParams{
				Layer1Client:      l1Client,
				CommitteeAPIHosts: config.NodeAPIURLs([]string{node}),
				N:                 uint16(len(node)), //nolint:gosec
				T:                 uint16(quorum),    //nolint:gosec
				OriginatorKeyPair: kp,
				Textout:           os.Stdout,
				PackageID:         packageID,
				StateMetadata:     *stateMetadata,
			}

			chainID, err := apilib.DeployChain(ctx, par, stateControllerAddress)
			log.Check(err)

			config.AddChain(chainName, chainID.String())

			activateChain(ctx, node, chainName, chainID)
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
