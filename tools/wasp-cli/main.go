// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/packages/wasp"
	"github.com/iotaledger/wasp/tools/wasp-cli/authentication"
	"github.com/iotaledger/wasp/tools/wasp-cli/chain"
	"github.com/iotaledger/wasp/tools/wasp-cli/completion"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/decode"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/metrics"
	"github.com/iotaledger/wasp/tools/wasp-cli/peering"
	"github.com/iotaledger/wasp/tools/wasp-cli/wallet"
)

var rootCmd = &cobra.Command{
	Version: wasp.Version,
	Use:     "wasp-cli",
	Short:   "wasp-cli is a command line tool for interacting with Wasp and its smart contracts.",
	Long: `wasp-cli is a command line tool for interacting with Wasp and its smart contracts.
NOTE: this is alpha software, only suitable for testing purposes.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		config.Read()
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help() //nolint:errcheck
	},
}

func init() {
	rootCmd.AddCommand(completion.CompletionCmd(rootCmd.Root().Name()))

	authentication.Init(rootCmd)
	log.Init(rootCmd)
	config.Init(rootCmd)
	wallet.Init(rootCmd)
	chain.Init(rootCmd)
	decode.Init(rootCmd)
	peering.Init(rootCmd)
	metrics.Init(rootCmd)
}

func main() {
	log.Check(rootCmd.Execute())
}
