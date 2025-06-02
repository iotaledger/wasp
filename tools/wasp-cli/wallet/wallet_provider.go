package wallet

import (
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func initWalletProviderCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "wallet-provider (keychain, sdk_ledger, sdk_stronghold)",
		Short: "Get or set wallet provider (keychain, sdk_ledger, sdk_stronghold)",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				log.Printf("Wallet provider: %s\n", string(wallet.GetWalletProvider()))
				return
			}

			log.Check(wallet.SetWalletProvider(wallet.WalletProvider(args[0])))
			log.Check(config.WriteConfig())
		},
	}
}
