package authentication

import (
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
)

func initAuthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "auth <command>",
		Short: "Authentication tools",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
}

func Init(rootCmd *cobra.Command) {
	authCmd := initAuthCmd()
	loginCmd := initLoginCmd()
	infoCmd := initInfoCmd()
	importCmd := initImportCmd()

	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(&cobra.Command{
		Use:        "login",
		Deprecated: "use 'auth login' instead",
	})
	authCmd.AddCommand(initSetTokenCmd())

	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(infoCmd)
	authCmd.AddCommand(importCmd)

	loginCmd.PersistentFlags().StringVarP(&username, "username", "u", "", "username")
	loginCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "password")
}
