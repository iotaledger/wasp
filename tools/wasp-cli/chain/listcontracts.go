package chain

import (
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func initListContractsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-contracts",
		Short: "List deployed contracts in chain",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			records, err := SCClient(root.Contract.Hname()).CallView(root.ViewGetContractRecords.Name, nil)
			log.Check(err)
			contracts, err := root.DecodeContractRegistry(collections.NewMapReadOnly(records, root.StateVarContractRegistry))
			log.Check(err)

			log.Printf("Total %d contracts in chain %s\n", len(contracts), GetCurrentChainID())

			header := []string{
				"hname",
				"name",
				"description",
				"proghash",
				"owner fee",
				"validator fee",
			}
			rows := make([][]string, len(contracts))
			i := 0
			for hname, c := range contracts {
				rows[i] = []string{
					hname.String(),
					c.Name,
					c.Description,
					c.ProgramHash.String(),
				}
				i++
			}
			log.PrintTable(header, rows)
		},
	}
}
