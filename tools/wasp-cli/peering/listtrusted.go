// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package peering

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/waspcmd"
)

func initListTrustedCmd() *cobra.Command {
	var node string

	cmd := &cobra.Command{
		Use:   "list-trusted",
		Short: "List trusted wasp nodes.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)

			ctx := context.Background()
			client := cliclients.WaspClientWithVersionCheck(ctx, node)
			trustedList, _, err := client.NodeAPI.GetTrustedPeers(ctx).Execute()
			log.Check(err)

			header := []string{"Name", "PubKey", "PeeringURL", "Trusted"}
			rows := make([][]string, len(trustedList))
			for i := range rows {
				rows[i] = []string{
					trustedList[i].Name,
					trustedList[i].PublicKey,
					trustedList[i].PeeringURL,
					fmt.Sprintf("%v", trustedList[i].IsTrusted),
				}
			}
			log.PrintTable(header, rows)
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)

	return cmd
}
