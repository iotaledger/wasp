// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package peering

import (
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/packages/log"
)

var peeringCmd = &cobra.Command{
	Use:   "peering <command>",
	Short: "Configure peering.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		log.Check(cmd.Help())
	},
}

func Init(rootCmd *cobra.Command) {
	rootCmd.AddCommand(peeringCmd)
	peeringCmd.AddCommand(infoCmd)
	peeringCmd.AddCommand(trustCmd)
	peeringCmd.AddCommand(distrustCmd)
	peeringCmd.AddCommand(listTrustedCmd)
}
