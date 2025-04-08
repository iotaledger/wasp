package setup

import (
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
)

func initRefreshL1ParamsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "refresh-l1-params",
		Short: "Refresh L1 params from node",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			config.RefreshL1ParamsFromNode()
		},
	}
}

func Init(rootCmd *cobra.Command) {
	rootCmd.PersistentFlags().StringVarP(&config.ConfigPath, "config", "c", "", "path to wasp-cli.json")
	rootCmd.PersistentFlags().StringVarP(&config.WaitForCompletion, "wait", "w", config.DefaultWaitForCompletion, "wait time for request completion, should not be less than 1 sec")

	rootCmd.AddCommand(initCheckVersionsCmd())
	rootCmd.AddCommand(initConfigSetCmd())
	rootCmd.AddCommand(initRefreshL1ParamsCmd())
}
