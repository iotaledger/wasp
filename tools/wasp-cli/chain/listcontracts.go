package chain

import (
	"github.com/spf13/cobra"
)

func initListContractsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "list-contracts",
		Short:      "List deployed contracts in chain",
		Deprecated: "no longer required",
	}

	return cmd
}
