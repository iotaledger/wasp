// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"fmt"
	"os"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/wallet"
	"github.com/spf13/cobra"
)

var rotateCmd = &cobra.Command{
	Use:   "rotate <new state controller address>",
	Short: "Issues a tx that changes the chain state controller",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		l1Client := config.L1Client()
		prefix, newStateControllerAddr, err := iotago.ParseBech32(args[0])
		log.Check(err)
		if parameters.L1().Protocol.Bech32HRP != prefix {
			log.Fatalf("unexpected prefix. expected: %s, actual: %s", parameters.L1().Protocol.Bech32HRP, prefix)
		}

		kp := wallet.Load().KeyPair

		chainID := *GetCurrentChainID().AsAliasID()

		chainOutputID, chainOutput, err := l1Client.GetAliasOutput(chainID)
		log.Check(err)

		tx, err := transaction.NewRotateChainStateControllerTx(
			chainID,
			newStateControllerAddr,
			chainOutputID,
			chainOutput,
			kp,
		)
		log.Check(err)

		err = l1Client.PostTx((tx))
		if err != nil {
			panic(err)
		}
		log.Check(err)

		txID, err := tx.ID()
		log.Check(err)
		fmt.Fprintf(os.Stdout, "chain rotation transaction issued successfully. TXID: %s", txID.ToHex())
	},
}
