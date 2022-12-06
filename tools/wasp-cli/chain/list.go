package chain

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List deployed chains",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		client := config.WaspClient(config.MustWaspAPI())
		chains, err := client.GetChainRecordList()
		log.Check(err)
		model := &ListChainModel{
			Length:  len(chains),
			BaseURL: client.BaseURL(),
		}
		model.Chains = make(map[string]bool)

		for _, chain := range chains {
			chainID := chain.ChainID()
			model.Chains[(&chainID).String()] = chain.Active
		}
		log.PrintCLIOutput(model)
		showChainList(chains)
	},
}

func showChainList(chains []*registry.ChainRecord) {
	header := []string{"chainid", "active"}
	rows := make([][]string, len(chains))
	for i, chain := range chains {
		chainID := chain.ChainID()
		rows[i] = []string{
			(&chainID).String(),
			fmt.Sprintf("%v", chain.Active),
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
