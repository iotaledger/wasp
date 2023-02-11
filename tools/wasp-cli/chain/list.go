package chain

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

func initListCmd() *cobra.Command {
	var node string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List deployed chains",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			client := cliclients.WaspClient(node)
			chains, _, err := client.ChainsApi.GetChains(context.Background()).Execute() //nolint:bodyclose // false positive
			log.Check(err)

			model := &ListChainModel{
				Length:  len(chains),
				BaseURL: client.GetConfig().Host,
			}

			model.Chains = make(map[string]bool)

			for _, chain := range chains {
				model.Chains[chain.ChainID] = chain.IsActive
			}

			log.PrintCLIOutput(model)
			showChainList(chains)
		},
	}
	waspcmd.WithWaspNodeFlag(cmd, &node)
	return cmd
}

func showChainList(chains []apiclient.ChainInfoResponse) {
	header := []string{"chainid", "active"}
	rows := make([][]string, len(chains))
	for i, chain := range chains {
		rows[i] = []string{
			chain.ChainID,
			fmt.Sprintf("%v", chain.IsActive),
		}
	}
	log.PrintTable(header, rows)
}

type ListChainModel struct {
	Length  int
	BaseURL string
	Chains  map[string]bool
}

var _ log.CLIOutput = &ListChainModel{}

func (l *ListChainModel) AsText() (string, error) {
	template := `Total {{ .Length }} chain(s) in wasp node {{ .BaseURL }}`
	return log.ParseCLIOutputTemplate(l, template)
}
