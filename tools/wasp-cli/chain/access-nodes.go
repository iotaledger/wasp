// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func initPermitionlessAccessNodesCmd() *cobra.Command {
	var nodes []int

	cmd := &cobra.Command{
		Use:   "access-nodes <action (add|remove)> <pubkey>",
		Short: "Changes the access nodes of a chain for the target node.",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if nodes == nil {
				nodes = GetAllWaspNodes()
			}
			chainID := GetCurrentChainID()
			action := args[0]
			pubKey := args[1]

			for _, i := range nodes {
				client := config.WaspClient(config.MustWaspAPI(i))
				switch action {
				case "add":
					err := client.AddAccessNode(chainID, pubKey)
					log.Check(err)
				case "remove":
					err := client.RemoveAccessNode(chainID, pubKey)
					log.Check(err)
				default:
					log.Fatalf("unknown action: %s", action)
				}
			}
		},
	}

	cmd.Flags().IntSliceVarP(&nodes, "nodes", "", nil, "wasp nodes to execute the command in (ex: 0,1,2,3) (default: all nodes)")

	return cmd
}
