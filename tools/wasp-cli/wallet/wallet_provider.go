package wallet

import (
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/wallet"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
	"github.com/spf13/cobra"
)

func initWalletProviderCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "provider (keychain, ledger)",
		Short: "Get or set wallet provider (keychain, ledger)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				log.Printf("Wallet provider: %s\n", string(wallet.GetWalletProvider()))
				return nil
			}

			if err := wallet.SetWalletProvider(wallet.WalletProvider(args[0])); err != nil {
				return err
			}
			if err := config.WriteConfig(); err != nil {
				return err
			}
			return nil
		},
	}
}
