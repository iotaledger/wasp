// Package wallet provides commands for managing and interacting with IOTA wallets,
// enabling users to perform various cryptocurrency operations.
package wallet

import (
	"github.com/spf13/cobra"
)

func Init(rootCmd *cobra.Command) {
	walletCmd := initWalletCmd()
	rootCmd.AddCommand(walletCmd)

	walletCmd.AddCommand(initInitCmd())
	walletCmd.AddCommand(initAddressCmd())
	walletCmd.AddCommand(initBalanceCmd())
	walletCmd.AddCommand(initSendFundsCmd())
	walletCmd.AddCommand(initRequestFundsCmd())
	walletCmd.AddCommand(initWalletProviderCmd())
	walletCmd.AddCommand(initMergeCmd())

	rootCmd.AddCommand(deprecated("init", "use 'wallet init' instead"))
	rootCmd.AddCommand(deprecated("address", "use 'wallet address' instead"))
	rootCmd.AddCommand(deprecated("balance", "use 'wallet balance' instead"))
	rootCmd.AddCommand(deprecated("send-funds", "use 'wallet send-funds' instead"))
	rootCmd.AddCommand(deprecated("request-funds", "use 'wallet request-funds' instead"))
	rootCmd.AddCommand(deprecated("wallet-provider", "use 'wallet provider' instead"))
	rootCmd.AddCommand(deprecated("wallet-migrate", "no longer supported"))
	rootCmd.AddCommand(deprecated("merge", "use 'wallet merge' instead"))
}

func deprecated(cmd, msg string) *cobra.Command {
	return &cobra.Command{
		Use:        cmd,
		Deprecated: msg,
	}
}

func initWalletCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "wallet <command>",
		Short: "Wallet tools",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
}
