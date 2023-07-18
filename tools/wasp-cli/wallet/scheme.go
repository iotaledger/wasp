package wallet

import (
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func initSchemeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "wallet-scheme (keychain, sdk_ledger, sdk_stronghold)",
		Short: "Get or set wallet scheme (keychain, sdk_ledger, sdk_stronghold)",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				log.Printf("Wallet scheme: %s\n", string(wallet.GetWalletScheme()))
				return
			}

			log.Check(wallet.SetWalletScheme(wallet.WalletScheme(args[0])))
		},
	}
}
