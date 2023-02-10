package init

import (
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
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

func Init(rootCmd *cobra.Command, waspVersion string) {
	rootCmd.PersistentFlags().StringVarP(&config.ConfigPath, "config", "c", "wasp-cli.json", "path to wasp-cli.json")
	rootCmd.PersistentFlags().BoolVarP(&config.WaitForCompletion, "wait", "w", true, "wait for request completion")

	rootCmd.AddCommand(initCheckVersionsCmd(waspVersion))
	rootCmd.AddCommand(initConfigSetCmd())
	rootCmd.AddCommand(initRefreshL1ParamsCmd())

	// The first time parameters.L1() is called, it will be initialized with this function
	parameters.InitL1Lazy(func() {
		cliclients.L1Client()

		if config.L1ParamsExpired() {
			config.RefreshL1ParamsFromNode()
		} else {
			config.LoadL1ParamsFromConfig()
		}
	})
}
