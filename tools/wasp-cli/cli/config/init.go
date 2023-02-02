package config

import (
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/packages/parameters"
)

func initConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			v := args[1]
			switch v {
			case "true":
				Set(args[0], true)
			case "false":
				Set(args[0], false)
			default:
				Set(args[0], v)
			}
		},
	}
}

func initRefreshL1ParamsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "refresh-l1-params",
		Short: "Refresh L1 params from node",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			refreshL1ParamsFromNode()
		},
	}
}

func Init(rootCmd *cobra.Command, waspVersion string) {
	rootCmd.PersistentFlags().StringVarP(&ConfigPath, "config", "c", "wasp-cli.json", "path to wasp-cli.json")
	rootCmd.PersistentFlags().BoolVarP(&WaitForCompletion, "wait", "w", true, "wait for request completion")

	rootCmd.AddCommand(initConfigSetCmd())
	rootCmd.AddCommand(initCheckVersionsCmd(waspVersion))
	rootCmd.AddCommand(initRefreshL1ParamsCmd())

	// The first time parameters.L1() is called, it will be initialized with this function
	parameters.InitL1Lazy(func() {
		if l1ParamsExpired() {
			refreshL1ParamsFromNode()
		} else {
			loadL1ParamsFromConfig()
		}
	})
}
