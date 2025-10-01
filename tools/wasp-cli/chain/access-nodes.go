// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package chain provides functionality for interacting with and managing IOTA smart contract chains.
package chain

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/waspcmd"
)

func initPermissionlessAccessNodesCmd() *cobra.Command {
	var node string
	var chain string
	var peers []string

	cmd := &cobra.Command{
		Use:   "access-nodes <action (add|remove)> --peers=<...>",
		Short: "Changes the access nodes of a chain for the target node.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			node, err = waspcmd.DefaultWaspNodeFallback(node)
			if err != nil {
				return err
			}
			chain, err = defaultChainFallback(chain)
			if err != nil {
				return err
			}

			action := args[0]
			node, err = waspcmd.DefaultWaspNodeFallback(node)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client := cliclients.WaspClientWithVersionCheck(ctx, node)

			var executeActionFunc func(peer string) error

			switch action {
			case "add":
				executeActionFunc = func(peer string) error {
					_, err := client.ChainsAPI.
						AddAccessNode(ctx, peer).
						Execute() //nolint:bodyclose // false positive
					if err != nil {
						return err
					}
					log.Printf("added %s as an access node\n", peer)
					return nil
				}
			case "remove":
				executeActionFunc = func(peer string) error {
					_, err := client.ChainsAPI.
						RemoveAccessNode(ctx, peer).
						Execute() //nolint:bodyclose // false positive
					if err != nil {
						return err
					}
					log.Printf("removed %s as an access node\n", peer)
					return nil
				}
			default:
				return fmt.Errorf("unknown action: %s", action)
			}

			for _, peer := range peers {
				if err := executeActionFunc(peer); err != nil {
					return err
				}
			}
			return nil
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chain)
	waspcmd.WithPeersFlag(cmd, &peers)

	return cmd
}
