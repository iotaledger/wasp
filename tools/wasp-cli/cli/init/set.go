package init

import (
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
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
				config.Set(args[0], true)
			case "false":
				config.Set(args[0], false)
			default:
				config.Set(args[0], v)
			}
		},
	}
}
