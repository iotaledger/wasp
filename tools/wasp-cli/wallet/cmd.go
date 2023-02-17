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

	rootCmd.PersistentFlags().IntVarP(&wallet.AddressIndex, "address-index", "i", 0, "address index")
}
