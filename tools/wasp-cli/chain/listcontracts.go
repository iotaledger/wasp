package chain

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/clients"
	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func initListContractsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-contracts",
		Short: "List deployed contracts in chain",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			client := cliclients.WaspClientForIndex()

			// TODO: Implement root view calls into v2 api
			records, _, err := client.RequestsApi.CallView(context.Background()).ContractCallViewRequest(apiclient.ContractCallViewRequest{
				ContractName: root.Contract.Name,
				FunctionName: root.ViewGetContractRecords.Name,
			}).Execute()

			if records == nil {
				log.Fatal("Could not fetch contract records")
			}

			parsedRecords, err := clients.APIJsonDictToDict(*records)
			log.Check(err)

			contracts, err := root.DecodeContractRegistry(collections.NewMapReadOnly(parsedRecords, root.StateVarContractRegistry))
			log.Check(err)

			log.Printf("Total %d contracts in chain %s\n", len(contracts), config.GetCurrentChainID())

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
