// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"strings"

	goversion "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/components/app"
	"github.com/iotaledger/wasp/tools/wasp-cli/authentication"
	"github.com/iotaledger/wasp/tools/wasp-cli/chain"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/setup"
	"github.com/iotaledger/wasp/tools/wasp-cli/decode"
	"github.com/iotaledger/wasp/tools/wasp-cli/disrec"
	"github.com/iotaledger/wasp/tools/wasp-cli/inspection"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/metrics"
	"github.com/iotaledger/wasp/tools/wasp-cli/peering"
	"github.com/iotaledger/wasp/tools/wasp-cli/wallet"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

var rootCmd *cobra.Command

func initRootCmd(waspVersion string) *cobra.Command {
	return &cobra.Command{
		Version: waspVersion,
		Use:     "wasp-cli",
		Short:   "wasp-cli is a command line tool for interacting with Wasp and its smart contracts.",
		Long: `wasp-cli is a command line tool for interacting with Wasp and its smart contracts.
	NOTE: this is alpha software, only suitable for testing purposes.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			config.Read()

			whitelistedCommands := map[string]struct{}{
				"init":            {},
				"wallet-migrate":  {},
				"wallet-provider": {},
			}

			_, ok := whitelistedCommands[cmd.Name()]

			if config.GetSeedForMigration() != "" && !ok {
				log.Printf("\n\nWarning\n\n")
				log.Printf("Wasp-cli is now utilizing the IOTA SDK and your OS Keychain to handle your seed more securely.\n")
				log.Printf("Therefore, seeds can not be stored inside the config file anymore.\n")
				log.Printf("Please run `wasp-cli wallet-migrate keychain` to move your seed into the Keychain of your operating system,\n")
				log.Printf("or switch to alternative wallet providers such as the Ledger with: `wasp-cli wallet-provider sdk_ledger`.\n")

				log.Fatalf("The cli will now exit.")
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help() //nolint:errcheck
		},
	}
}

func init() {
	waspVersion := app.Version

	if strings.HasPrefix(strings.ToLower(waspVersion), "v") {
		if _, err := goversion.NewSemver(waspVersion[1:]); err == nil {
			// version is a valid SemVer with a "v" prefix => remove the "v" prefix
			waspVersion = waspVersion[1:]
		}
	}

	rootCmd = initRootCmd(waspVersion)
	rootCmd.PersistentFlags().BoolVar(&cliclients.SkipCheckVersions, "skip-version-check", true, "skip-version-check")

	log.Init(rootCmd)
	rootCmd.AddCommand(&cobra.Command{
		Use:        "gen_completion",
		Deprecated: "use 'completion' instead",
	})
	setup.Init(rootCmd)
	authentication.Init(rootCmd)
	waspcmd.Init(rootCmd)
	wallet.Init(rootCmd)
	chain.Init(rootCmd)
	decode.Init(rootCmd)
	peering.Init(rootCmd)
	metrics.Init(rootCmd)
	disrec.Init(rootCmd)
	inspection.Init(rootCmd)
}

func main() {
	log.Check(rootCmd.Execute())
}
