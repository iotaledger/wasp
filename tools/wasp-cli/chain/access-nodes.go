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
	var peers []string

	cmd := &cobra.Command{
		Use:   "access-nodes <action (add|remove)> --peers=<...>",
		Short: "Changes the access nodes of a chain for the target node.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			chain = defaultChainFallback(chain)

			chainID := config.GetChain(chain)
			action := args[0]
			node = waspcmd.DefaultWaspNodeFallback(node)

			client := cliclients.WaspClient(node)

			var executeActionFunc func(peer string)

			switch action {
			case "add":
				executeActionFunc = func(peer string) {
					_, err := client.ChainsApi.
						AddAccessNode(context.Background(), chainID.String(), peer).
						Execute() //nolint:bodyclose // false positive
					log.Check(err)
					log.Printf("added %s as an access node\n", peer)
				}
			case "remove":
				executeActionFunc = func(peer string) {
					_, err := client.ChainsApi.
						RemoveAccessNode(context.Background(), chainID.String(), peer).
						Execute() //nolint:bodyclose // false positive
					log.Check(err)
					log.Printf("removed %s as an access node\n", peer)
				}
			default:
				log.Fatalf("unknown action: %s", action)
			}

			for _, peer := range peers {
				executeActionFunc(peer)
			}
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chain)
	waspcmd.WithPeersFlag(cmd, &peers)

	return cmd
}
