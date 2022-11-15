// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package peering

import (
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

var listTrustedCmd = &cobra.Command{
	Use:   "list-trusted",
	Short: "List trusted wasp nodes.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		trustedList, err := config.WaspClient(config.MustWaspAPI()).GetPeeringTrustedList()
		log.Check(err)
		header := []string{"PubKey", "NetID"}
		rows := make([][]string, len(trustedList))
		for i := range rows {
			rows[i] = []string{
				trustedList[i].PubKey,
				trustedList[i].NetID,
			}
		}
		log.PrintTable(header, rows)
	},
}
