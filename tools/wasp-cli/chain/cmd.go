package chain

import (
	"github.com/spf13/cobra"
)

func initChainCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "chain <command>",
		Short: "Interact with a chain",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
}

func Init(rootCmd *cobra.Command) {
	chainCmd := initChainCmd()
	rootCmd.AddCommand(chainCmd)

	chainCmd.AddCommand(initDeployCmd())
	chainCmd.AddCommand(initImportCmd())
	chainCmd.AddCommand(initInfoCmd())
	chainCmd.AddCommand(initListContractsCmd())
	chainCmd.AddCommand(initDeployMoveContractCmd())
	chainCmd.AddCommand(initBalanceCmd())
	chainCmd.AddCommand(initAccountObjectsCmd())
	chainCmd.AddCommand(initDepositCmd())
	chainCmd.AddCommand(initBlockCmd())
	chainCmd.AddCommand(initRequestCmd())
	chainCmd.AddCommand(initPostRequestCmd())
	chainCmd.AddCommand(initCallViewCmd())
	chainCmd.AddCommand(initActivateCmd())
	chainCmd.AddCommand(initDeactivateCmd())
	chainCmd.AddCommand(initRunDKGCmd())
	chainCmd.AddCommand(initRotateCmd())
	chainCmd.AddCommand(initChangeGovControllerCmd())
	chainCmd.AddCommand(initChangeAccessNodesCmd())
	chainCmd.AddCommand(initDisableFeePolicyCmd())
	chainCmd.AddCommand(initPermissionlessAccessNodesCmd())
	chainCmd.AddCommand(initAddChainCmd())
	chainCmd.AddCommand(initRegisterERC20NativeTokenCmd())
	chainCmd.AddCommand(initSetCoinMetadataCmd())
	chainCmd.AddCommand(initMetadataCmd())
	chainCmd.AddCommand(initBuildIndex())
}
