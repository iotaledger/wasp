// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/cobra"
)

func changeAccessNodesCmd() *cobra.Command {
	var offLedger bool

	cmd := &cobra.Command{
		Use:   "change-access-nodes <action (accept|remove|drop)> <pubkey>",
		Short: "Changes the access nodes of a chain.",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args)%2 != 0 {
				log.Fatalf("wrong number of arguments")
			}
			pars := governance.NewChangeAccessNodesRequest()
			for i := 1; i < len(args); i += 2 {
				action := args[i-1]
				pubkey, err := cryptolib.NewPublicKeyFromString(args[i])
				log.Check(err)
				switch action {
				case "accept":
					pars.Accept(pubkey)
				case "remove":
					pars.Remove(pubkey)
				case "drop":
					pars.Drop(pubkey)
				}
			}
			params := chainclient.PostRequestParams{
				Args: pars.AsDict(),
			}
			postRequest(
				governance.Contract.Name,
				governance.FuncChangeAccessNodes.Name,
				params,
				offLedger)
		},
	}

	cmd.Flags().BoolVarP(&offLedger, "off-ledger", "o", false,
		"post an off-ledger request",
	)

	return cmd
}
