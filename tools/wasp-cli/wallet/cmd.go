package wallet

import (
	"github.com/spf13/cobra"
)

func Init(rootCmd *cobra.Command) {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addressCmd)
	rootCmd.AddCommand(balanceCmd)
	rootCmd.AddCommand(mintCmd)
	rootCmd.AddCommand(sendFundsCmd())
	rootCmd.AddCommand(requestFundsCmd)

	rootCmd.PersistentFlags().IntVarP(&addressIndex, "address-index", "i", 0, "address index")
}
