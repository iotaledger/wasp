// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package peering

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func initListTrustedCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-trusted",
		Short: "List trusted wasp nodes.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			client := cliclients.WaspClient()
			trustedList, _, err := client.NodeApi.GetTrustedPeers(context.Background()).Execute()
			log.Check(err)

			header := []string{"PubKey", "NetID"}
			rows := make([][]string, len(trustedList))
			for i := range rows {
				rows[i] = []string{
					trustedList[i].PublicKey,
					trustedList[i].NetId,
				}
			}
			log.PrintTable(header, rows)
		},
	}
}
