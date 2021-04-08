package chain

import (
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/cobra"
)

var activateCmd = &cobra.Command{
	Use:   "activate",
	Short: "Activate the chain",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		log.Check(MultiClient().ActivateChain(*GetCurrentChainID()))
	},
}

var deactivateCmd = &cobra.Command{
	Use:   "deactivate",
	Short: "Deactivate the chain",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		log.Check(MultiClient().DeactivateChain(*GetCurrentChainID()))
	},
}
