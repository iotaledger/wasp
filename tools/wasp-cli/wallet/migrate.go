package wallet

import (
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet"
)

func initMigrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "wallet-migrate (keychain, sdk_stronghold)",
		Short: "Migrates a seed inside `wasp-cli.json` to a certain wallet provider (keychain, sdk_stronghold)",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			wallet.Migrate(wallet.WalletScheme(args[0]))
		},
	}
}
