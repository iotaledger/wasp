package chain

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/packages/log"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List deployed chains",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		client := config.WaspClient()
		chains, err := client.GetChainRecordList()
		log.Check(err)
		log.Printf("Total %d chain(s) in wasp node %s\n", len(chains), client.BaseURL())
		showChainList(chains)
	},
}

func showChainList(chains []*registry.ChainRecord) {
	header := []string{"chainid", "active"}
	rows := make([][]string, len(chains))
	for i, chain := range chains {
		rows[i] = []string{
			chain.ChainID.String(),
			fmt.Sprintf("%v", chain.Active),
		}
	}
	log.PrintTable(header, rows)
}
