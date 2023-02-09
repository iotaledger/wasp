// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package peering

import (
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func initPeeringCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "peering <command>",
		Short: "Configure peering.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			log.Check(cmd.Help())
		},
	}
}

func Init(rootCmd *cobra.Command) {
	peeringCmd := initPeeringCmd()
	rootCmd.AddCommand(peeringCmd)
	peeringCmd.AddCommand(initInfoCmd())
	peeringCmd.AddCommand(initTrustCmd())
	peeringCmd.AddCommand(initDistrustCmd())
	peeringCmd.AddCommand(initListTrustedCmd())
	peeringCmd.AddCommand(initImportTrustedJSONCmd())
}
