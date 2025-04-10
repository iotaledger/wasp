// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

func initRotateCmd() *cobra.Command {
	var (
		node  string
		chain string
	)
	cmd := &cobra.Command{
		Use:   "rotate <new committee address>",
		Short: "Ask this node to propose rotation address.",
		Long:  "Empty or missing argument means we cancel attempt to rotate the chain.",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			ctx := context.Background()
			client := cliclients.WaspClientWithVersionCheck(ctx, node)

			var rotateToAddress *string
			if len(args) == 1 && args[0] != "" {
				rotateToAddress = &args[0]
			}

			_, err := client.ChainsAPI.RotateChain(ctx).RotateRequest(apiclient.RotateChainRequest{
				RotateToAddress: rotateToAddress,
			}).Execute() //nolint:bodyclose // false positive
			log.Check(err)
			if rotateToAddress == nil {
				log.Printf("Will stop proposing chain rotation to another address.\n")
			} else {
				log.Printf("Will attempt to rotate to %s\n", *rotateToAddress)
			}
		},
	}
	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chain)
	return cmd
}

//lint:ignore U1000 Ignore unused function temporarily for debugging
func setMaintenanceStatus(ctx context.Context, client *apiclient.APIClient, chain string, status bool, offledger bool) {
	msg := governance.FuncStartMaintenance.Message()
	if !status {
		msg = governance.FuncStopMaintenance.Message()
	}
	postRequest(ctx, client, chain, msg, chainclient.PostRequestParams{
		GasBudget: iotaclient.DefaultGasBudget,
	}, offledger)
}

func initChangeGovControllerCmd() *cobra.Command {
	var chain string

	cmd := &cobra.Command{
		Use:   "change-gov-controller <address> --chain=<chainID>",
		Short: "Changes the governance controller for a given chain (WARNING: you will lose control over the chain)",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			panic("refactor me: l1connection.OutputMap")
			/*
				chain := config.GetChain(defaultChainFallback(chain))

				_, newGovController, err := iotago.ParseBech32(args[0])
				log.Check(err)

				myWallet := wallet.Load()

				//outputSet, err := client.OutputMap(myWallet.Address())
				var outputSet iotago.OutputSet
				err = errors.New("refactor me: initChangeGovControllerCmd")
				log.Check(err)

				tx, err := transaction.NewChangeGovControllerTx(
					chain.AsAliasID(),
					newGovController,
					outputSet,
					myWallet,
				)
				_ = tx
				log.Check(err)

				panic("refactor me: l1connection.PostTxAndWaitUntilConfirmation")

				log.Check(err)*/
		},
	}

	withChainFlag(cmd, &chain)
	return cmd
}
