// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	cliwallet "github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

func initRotateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rotate <new state controller address>",
		Short: "Issues a tx that changes the chain state controller",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			prefix, newStateControllerAddr, err := iotago.ParseBech32(args[0])
			log.Check(err)
			if parameters.L1().Protocol.Bech32HRP != prefix {
				log.Fatalf("unexpected prefix. expected: %s, actual: %s", parameters.L1().Protocol.Bech32HRP, prefix)
			}
			rotateTo(newStateControllerAddr)
		},
	}
}

func initRotateWithDKGCmd() *cobra.Command {
	var (
		nodes  []string
		quorum int
	)

	cmd := &cobra.Command{
		Use:   "rotate-with-dkg --nodes=<...>",
		Short: "Runs the DKG on the selected nodes, then issues a tx that changes the chain state controller",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if len(nodes) == 0 {
				nodes = []string{config.MustGetDefaultWaspNode()}
			}
			controllerAddr := doDKG(nodes, quorum)
			rotateTo(controllerAddr)
		},
	}

	waspcmd.WithWaspNodesFlag(cmd, &nodes)
	log.Check(cmd.MarkFlagRequired("nodes"))
	cmd.Flags().IntVarP(&quorum, "quorum", "", 0, "quorum (default: 3/4s of the number of committee nodes)")
	return cmd
}

func rotateTo(newStateControllerAddr iotago.Address) {
	l1Client := cliclients.L1Client()

	wallet := cliwallet.Load()
	aliasID := config.GetCurrentChainID().AsAliasID()

	chainOutputID, chainOutput, err := l1Client.GetAliasOutput(aliasID)
	log.Check(err)

	tx, err := transaction.NewRotateChainStateControllerTx(
		aliasID,
		newStateControllerAddr,
		chainOutputID,
		chainOutput,
		wallet.KeyPair,
	)
	log.Check(err)
	log.Verbosef("issuing rotation tx, signed for address: %s", wallet.KeyPair.Address().Bech32(parameters.L1().Protocol.Bech32HRP))
	json, err := tx.MarshalJSON()
	log.Check(err)
	log.Verbosef("rotation tx: %s", string(json))

	_, err = l1Client.PostTxAndWaitUntilConfirmation(tx)
	if err != nil {
		panic(err)
	}
	log.Check(err)

	txID, err := tx.ID()
	log.Check(err)
	fmt.Fprintf(os.Stdout, "chain rotation transaction issued successfully. TXID: %s", txID.ToHex())
}
