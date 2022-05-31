package chain

import (
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/cobra"
)

var listContractsCmd = &cobra.Command{
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
			"creator",
			"owner fee",
			"validator fee",
		}
		rows := make([][]string, len(contracts))
		i := 0
		for hname, c := range contracts {
			creator := ""
			if c.HasCreator() {
				creator = c.Creator.String()
			}

			rows[i] = []string{
				hname.String(),
				c.Name,
				c.Description,
				c.ProgramHash.String(),
				creator,
			}
			i++
		}
		log.PrintTable(header, rows)
	},
}
