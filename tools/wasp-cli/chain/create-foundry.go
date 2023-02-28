package chain

import (
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/spf13/cobra"
)

func initCreateFoundryCmd() *cobra.Command {
	return buildPostRequestCmd(
		"create-foundry",
		"Call accounts core contract foundryCreateNew to create a new foundry",
		accounts.Contract.Name,
		accounts.FuncFoundryCreateNew.Name,
		func(cmd *cobra.Command) {
			cmd.Flags().String("token-scheme", "", "Token scheme")
		},
		func(cmd *cobra.Command) []string {
			return []string{"string", "t", "bytes", cmd.Flag("token-scheme").Value.String()}
		},
	)
}
