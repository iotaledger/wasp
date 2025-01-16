// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

func initRotateCmd() *cobra.Command {
	var chain string
	cmd := &cobra.Command{
		Use:   "rotate <new state controller address>",
		Short: "Issues a tx that changes the chain state controller",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			chain = defaultChainFallback(chain)

			newStateControllerAddr, err := cryptolib.NewAddressFromHexString(args[0])
			log.Check(err)

			rotateTo(chain, newStateControllerAddr)
		},
	}
	withChainFlag(cmd, &chain)
	return cmd
}

func initRotateWithDKGCmd() *cobra.Command {
	var (
		node            string
		peers           []string
		quorum          int
		chain           string
		skipMaintenance bool
		offLedger       bool
	)

	cmd := &cobra.Command{
		Use:   "rotate-with-dkg --peers=<...>",
		Short: "Runs the DKG on the selected peers, then issues a tx that changes the chain state controller",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			chain = defaultChainFallback(chain)
			node = waspcmd.DefaultWaspNodeFallback(node)
			ctx := context.Background()
			client := cliclients.WaspClientWithVersionCheck(ctx, node)

			if !skipMaintenance {
				setMaintenanceStatus(ctx, client, chain, true, offLedger)
				defer setMaintenanceStatus(ctx, client, chain, false, offLedger)
			}

			controllerAddr := doDKG(ctx, node, peers, quorum)
			rotateTo(chain, controllerAddr)
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)
	waspcmd.WithPeersFlag(cmd, &peers)
	withChainFlag(cmd, &chain)
	cmd.Flags().IntVarP(&quorum, "quorum", "", 0, "quorum (default: 3/4s of the number of committee nodes)")
	cmd.Flags().BoolVar(&skipMaintenance, "skip-maintenance", false, "quorum (default: 3/4s of the number of committee nodes)")
	cmd.Flags().BoolVarP(&offLedger, "off-ledger", "o", true,
		"post an off-ledger request",
	)

	return cmd
}

func rotateTo(chain string, newStateControllerAddr *cryptolib.Address) {
	panic("refactor me: l1Client.GetAliasOutput")
	/*
		myWallet := wallet.Load()
		aliasID := config.GetChain(chain).AsObjectID()

		var chainOutputID iotago.OutputID
		var chainOutput iotago.Output
		err := errors.New("refactor me: rotateTo")
		// chainOutputID, chainOutput, err := l1Client.GetAliasOutput(aliasID)
		log.Check(err)

		tx, err := transaction.NewRotateChainStateControllerTx(
			aliasID,
			newStateControllerAddr,
			chainOutputID,
			chainOutput,
			myWallet,
		)
		log.Check(err)

		// debug logging
		if log.DebugFlag {
			s, err2 := chainOutput.MarshalJSON()
			log.Check(err2)
			minSD := parameters.L1().Protocol.RentStructure.MinRent(chainOutput)
			log.Printf("original chain output: %s, minSD: %d\n", s, minSD)

			rotOut := tx.Essence.Outputs[0]
			s, err2 = rotOut.MarshalJSON()
			log.Check(err2)
			minSD = parameters.L1().Protocol.RentStructure.MinRent(rotOut)
			log.Printf("new chain output: %s, minSD: %d\n", s, minSD)

			json, err2 := tx.MarshalJSON()
			log.Check(err2)
			log.Printf("issuing rotation tx, signed for address: %s", myWallet.Address().String())
			log.Printf("rotation tx: %s", string(json))
		}

		panic("refactor me: l1Client.PostTxAndWaitUntilConfirmation")

		log.Check(err)

		txID, err := tx.ID()
		log.Check(err)
		fmt.Fprintf(os.Stdout, "Chain rotation transaction issued successfully.\nTXID: %s\n", txID.ToHex())

	*/
}

func setMaintenanceStatus(ctx context.Context, client *apiclient.APIClient, chain string, status bool, offledger bool) {
	msg := governance.FuncStartMaintenance.Message()
	if !status {
		msg = governance.FuncStopMaintenance.Message()
	}
	postRequest(ctx, client, chain, msg, chainclient.PostRequestParams{}, offledger, true)
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
