package wallet

import (
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet"
)

func Init(rootCmd *cobra.Command) {
	rootCmd.AddCommand(initInitCmd())
	rootCmd.AddCommand(initAddressCmd())
	rootCmd.AddCommand(initBalanceCmd())
	rootCmd.AddCommand(initSendFundsCmd())
	rootCmd.AddCommand(initRequestFundsCmd())
	rootCmd.AddCommand(initWalletProviderCmd())
	rootCmd.AddCommand(initMigrateCmd())

	rootCmd.PersistentFlags().Uint32VarP(&wallet.AddressIndex, "address-index", "i", 0, "address index")
}
