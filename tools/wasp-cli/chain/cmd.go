package chain

import (
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/cobra"
)

var chainCmd = &cobra.Command{
	Use:   "chain <command>",
	Short: "Interact with a chain",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		log.Check(cmd.Help())
	},
}

var plugins []func(*cobra.Command)

func Init(rootCmd *cobra.Command) {
	rootCmd.AddCommand(chainCmd)

	initAliasFlags(chainCmd)
	initUploadFlags(chainCmd)

	chainCmd.AddCommand(listCmd)
	chainCmd.AddCommand(deployCmd())
	chainCmd.AddCommand(infoCmd)
	chainCmd.AddCommand(listContractsCmd)
	chainCmd.AddCommand(deployContractCmd)
	chainCmd.AddCommand(listAccountsCmd)
	chainCmd.AddCommand(balanceCmd)
	chainCmd.AddCommand(depositCmd)
	chainCmd.AddCommand(listBlobsCmd)
	chainCmd.AddCommand(storeBlobCmd)
	chainCmd.AddCommand(showBlobCmd)
	chainCmd.AddCommand(logCmd)
	chainCmd.AddCommand(blockCmd())
	chainCmd.AddCommand(requestCmd())
	chainCmd.AddCommand(postRequestCmd())
	chainCmd.AddCommand(callViewCmd)
	chainCmd.AddCommand(activateCmd)
	chainCmd.AddCommand(deactivateCmd)

	for _, p := range plugins {
		p(chainCmd)
	}
}
