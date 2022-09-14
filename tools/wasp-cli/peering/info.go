// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package peering

import (
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Node info.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		info, err := config.WaspClient().GetPeeringSelf()
		log.Check(err)
		log.Printf("PubKey: %v\n", info.PubKey)
		log.Printf("NetID:  %v\n", info.NetID)
	},
}
