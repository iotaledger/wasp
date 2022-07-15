package chain

import "github.com/spf13/cobra"

var addChainCmd = &cobra.Command{
	Use:   "add <name> <chain id>",
	Short: "adds a chain to the list of chains",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		AddChainAlias(args[0], args[1])
	},
}
