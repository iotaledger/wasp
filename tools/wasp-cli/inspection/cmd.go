// Package inspection provides tools for inspecting and analyzing various
// aspects of the IOTA smart contract system through wasp-cli.
package inspection

import (
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
)

func initInspect() *cobra.Command {
	return &cobra.Command{
		Use:   "inspect <command>",
		Short: "Get information about a given object",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			log.Check(cmd.Help())
		},
	}
}

func Init(rootCmd *cobra.Command) {
	inspectCmd := initInspect()
	rootCmd.AddCommand(inspectCmd)

	inspectCmd.AddCommand(initAssetsBagCmd())
	inspectCmd.AddCommand(initAnchorCmd())
	inspectCmd.AddCommand(initRequestsCmd())
}
