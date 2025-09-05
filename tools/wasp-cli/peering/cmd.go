// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package peering implements functionality for managing peer connections
// between Wasp nodes in the IOTA smart contract network.
package peering

import (
	"github.com/spf13/cobra"
)

func initPeeringCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "peering <command>",
		Short: "Configure peering.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
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
	peeringCmd.AddCommand(initExportTrustedJSONCmd())
	peeringCmd.AddCommand(initImportTrustedJSONCmd())
}
