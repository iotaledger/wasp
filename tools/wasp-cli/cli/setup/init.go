package setup

import (
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
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
	rootCmd.PersistentFlags().BoolVarP(&config.WaitForCompletion, "wait", "w", true, "wait for request completion")

	rootCmd.AddCommand(initCheckVersionsCmd())
	rootCmd.AddCommand(initConfigSetCmd())
	rootCmd.AddCommand(initRefreshL1ParamsCmd())

	client := iotaclient.NewHTTP(config.L1APIAddress(), iotaclient.WaitForEffectsDisabled)
	err := parameters.InitL1(*client, log.HiveLogger())
	if err != nil {
		panic(err)
	}
}
