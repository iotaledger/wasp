// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/iotaledger/hive.go/kvstore/mapdb"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
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
	cliutil "github.com/iotaledger/wasp/tools/wasp-cli/util"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

func initializeMigrateChainState(stateController *cryptolib.Address, gasCoinObject iotago.ObjectID, anchor *iscmove.AnchorWithRef) *transaction.StateMetadata {
	initParams := origin.DefaultInitParams(isc.NewAddressAgentID(stateController)).Encode()
	store := indexedstore.New(state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB()))
	_, stateMetadata := origin.InitChain(allmigrations.LatestSchemaVersion, store, initParams, gasCoinObject, isc.GasCoinTargetValue, isc.BaseTokenCoinInfo)
	return stateMetadata
}

func initMigrateDeployCmd() *cobra.Command {
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
		Use:   "migrate-deploy --chain=<name>",
		Short: "Migrates a Stardust chain to Rebased",
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
			packageID := config.GetPackageID()
			stateControllerAddress := doDKG(ctx, node, peers, quorum)

			/**
			Stardust -> Rebased migration

			1) Use the TypeScript L1 AliasOutput migration tool, which prints unsigned, hex-encoded transaction data.
			2) Execute the transaction right here using the wallet which owns the AliasOutput (Governor Address)
			3) Receive the AssetsBag after the transaction has been executed.
			4) Store the AssetsBagRef
			5) CreateAndSendGasCoin
			6) CreateAnchorWithAssetsBagRef(assetsBagRef)
			7) Migrate the StateDB using the ChainID returned from CreateAnchorWithAssetsBagRef
			8) Extract the StateMetadata out of the StateDB Migration
			9) Update the Anchor using UpdateAnchorStateMetadata
			10) Proceed like a normal deployment (Activate chain on all nodes)
			*/

			gasCoin, err := cliutil.CreateAndSendGasCoin(ctx, l1Client, kp, stateControllerAddress.AsIotaAddress())
			log.Check(err)

			anchor, err := l1Client.L2().CreateAnchorWithAssetsBagRef(ctx, &iscmoveclient.CreateAnchorWithAssetsBagRefRequest{})

			fmt.Println(anchor.ObjectID) // <-- ChainID

			// Here we collect the migrated state
			stateMetadata := initializeMigrateChainState(stateControllerAddress, gasCoin, anchor)

			success, err := l1Client.L2().UpdateAnchorStateMetadata(ctx, &iscmoveclient.UpdateAnchorStateMetadataRequest{
				AnchorRef:     &anchor.ObjectRef,
				StateMetadata: stateMetadata.Bytes(),
			})

			if !success || err != nil {
				panic("omg")
			}

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
