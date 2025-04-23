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

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/origin"
	"github.com/iotaledger/wasp/packages/parameters"
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

func initializeNewChainState(chainAdmin *cryptolib.Address, gasCoinObject iotago.ObjectID, l1Params *parameters.L1Params) *transaction.StateMetadata {
	initParams := origin.DefaultInitParams(isc.NewAddressAgentID(chainAdmin)).Encode()
	store := indexedstore.New(state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB()))
	_, stateMetadata := origin.InitChain(allmigrations.LatestSchemaVersion, store, initParams, gasCoinObject, isc.GasCoinTargetValue, l1Params)
	return stateMetadata
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

			client := cliclients.WaspClientWithVersionCheck(ctx, node)
			_, header, err := client.ChainsAPI.GetChainInfo(ctx).Execute()
			defer header.Body.Close()

			// We expect a 404 if no chain has been deployed yet. In any other case, show the error.
			if err != nil && !strings.Contains(err.Error(), strconv.Itoa(http.StatusNotFound)) {
				log.Fatal(fmt.Errorf("failed to get current chain info: %w", err))
			}

			// Now check if the response is 404, if not, a Chain has already been deployed. Exit early.
			if header != nil && header.StatusCode != http.StatusNotFound {
				log.Fatal("A chain has already been deployed.")
			}

			packageID := config.GetPackageID()

			committeeAddr := doDKG(ctx, node, peers, quorum)

			l1Params, err := parameters.FetchLatest(context.Background(), l1Client.IotaClient())
			log.Check(err)

			gasCoin, err := cliutil.CreateAndSendGasCoin(ctx, l1Client, kp, committeeAddr.AsIotaAddress(), l1Params)
			log.Check(err)

			stateMetadata := initializeNewChainState(kp.Address(), gasCoin, l1Params)

			par := apilib.CreateChainParams{
				Layer1Client:      l1Client,
				CommitteeAPIHosts: config.NodeAPIURLs([]string{node}),
				Signer:            kp,
				Textout:           os.Stdout,
				PackageID:         packageID,
				StateMetadata:     *stateMetadata,
			}

			chainID, err := apilib.DeployChain(ctx, par, committeeAddr)
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
