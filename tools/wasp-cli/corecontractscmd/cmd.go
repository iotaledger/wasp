package corecontractscmd

import (
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/cobra"
)

var node, chain string

func Init(rootCmd *cobra.Command) {
	callCmd := initCallCmd()
	rootCmd.AddCommand(callCmd)

	callCmd.AddCommand(initEvmCmd())
}

func initCallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "call <contract>",
		Short: "Call a core contract entry point",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			log.Check(cmd.Help())
		},
	}

	cmd.PersistentFlags().StringVar(&chain, "chain", "", "Chain to use")
	cmd.PersistentFlags().StringVar(&node, "node", "", "Node to use")
	return cmd
}
