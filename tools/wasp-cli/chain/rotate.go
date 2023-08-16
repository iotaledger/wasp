// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet"
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

			prefix, newStateControllerAddr, err := iotago.ParseBech32(args[0])
			log.Check(err)
			if parameters.L1().Protocol.Bech32HRP != prefix {
				log.Fatalf("unexpected prefix. expected: %s, actual: %s", parameters.L1().Protocol.Bech32HRP, prefix)
			}
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

			if !skipMaintenance {
				setMaintenanceStatus(chain, node, true, offLedger)
				defer setMaintenanceStatus(chain, node, false, offLedger)
			}

			controllerAddr := doDKG(node, peers, quorum)
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

func rotateTo(chain string, newStateControllerAddr iotago.Address) {
	l1Client := cliclients.L1Client()

	myWallet := wallet.Load()
	aliasID := config.GetChain(chain).AsAliasID()

	chainOutputID, chainOutput, err := l1Client.GetAliasOutput(aliasID)
	log.Check(err)

	tx, err := transaction.NewRotateChainStateControllerTx(
		aliasID,
		newStateControllerAddr,
		chainOutputID,
		chainOutput,
		myWallet.KeyPair,
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
		log.Printf("issuing rotation tx, signed for address: %s", myWallet.KeyPair.Address().Bech32(parameters.L1().Protocol.Bech32HRP))
		log.Printf("rotation tx: %s", string(json))
	}

	_, err = l1Client.PostTxAndWaitUntilConfirmation(tx)
	if err != nil {
		panic(err)
	}
	log.Check(err)

	txID, err := tx.ID()
	log.Check(err)
	fmt.Fprintf(os.Stdout, "Chain rotation transaction issued successfully.\nTXID: %s\n", txID.ToHex())
}

func setMaintenanceStatus(chain, node string, status bool, offledger bool) {
	entrypoint := governance.FuncStartMaintenance.Name
	if !status {
		entrypoint = governance.FuncStopMaintenance.Name
	}
	params := chainclient.PostRequestParams{}
	postRequest(
		node,
		chain,
		governance.Contract.Name,
		entrypoint,
		params,
		offledger,
		true,
	)
}
