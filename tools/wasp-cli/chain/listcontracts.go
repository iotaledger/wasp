package chain

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

func initListContractsCmd() *cobra.Command {
	var node string
	var chain string

	cmd := &cobra.Command{
		Use:   "list-contracts",
		Short: "List deployed contracts in chain",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			chain = defaultChainFallback(chain)

			client := cliclients.WaspClient(node)
			contracts, _, err := client.ChainsApi.
				GetContracts(context.Background(), config.GetChain(chain).String()).
				Execute()

			log.Check(err)

			log.Printf("Total %d contracts in chain %s\n", len(contracts), config.GetChain(chain))

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
			for _, contract := range contracts {
				rows[i] = []string{
					contract.HName,
					contract.Name,
					contract.Description,
					contract.ProgramHash,
				}
				i++
			}
			log.PrintTable(header, rows)
		},
	}
	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chain)
	return cmd
}
