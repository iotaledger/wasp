// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

func initPermitionlessAccessNodesCmd() *cobra.Command {
	var node string
	var chain string

	cmd := &cobra.Command{
		Use:   "access-nodes <action (add|remove)> <pubkey>",
		Short: "Changes the access nodes of a chain for the target node.",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			chain = defaultChainFallback(chain)

			chainID := config.GetChain(chain)
			action := args[0]
			pubKey := args[1]
			node = waspcmd.DefaultWaspNodeFallback(node)

			client := cliclients.WaspClient(node)
			switch action {
			case "add":
				_, err := client.ChainsApi.
					AddAccessNode(context.Background(), chainID.String(), pubKey).
					Execute()
				log.Check(err)
			case "remove":
				_, err := client.ChainsApi.
					RemoveAccessNode(context.Background(), chainID.String(), pubKey).
					Execute()
				log.Check(err)
			default:
				log.Fatalf("unknown action: %s", action)
			}
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chain)

	return cmd
}
