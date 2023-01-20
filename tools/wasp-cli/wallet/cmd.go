package wallet

import (
	"github.com/spf13/cobra"
)

func Init(rootCmd *cobra.Command) {
	rootCmd.AddCommand(initInitCmd())
	rootCmd.AddCommand(initAddressCmd())
	rootCmd.AddCommand(initBalanceCmd())
	rootCmd.AddCommand(initSendFundsCmd())
	rootCmd.AddCommand(initRequestFundsCmd())

	rootCmd.PersistentFlags().IntVarP(&addressIndex, "address-index", "i", 0, "address index")
}
