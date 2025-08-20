package chain

import (
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/config"
)

func initAddChainCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <name> <chain id>",
		Short: "adds a chain to the list of chains",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			config.AddChain(args[0], args[1])
		},
	}
}
