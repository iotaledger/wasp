package chain

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func listCmd(args []string) {
	client := config.WaspClient()
	chains, err := client.GetChainRecordList()
	log.Check(err)
	log.Printf("Total %d chain(s) in wasp node %s\n", len(chains), client.BaseURL())
	showChainList(chains)
}

func showChainList(chains []*registry.ChainRecord) {
	header := []string{"chainid", "color", "committee", "active"}
	rows := make([][]string, len(chains))
	for i, chain := range chains {
		rows[i] = []string{
			chain.ChainID.String(),
			chain.Color.String(),
			fmt.Sprintf("%v", chain.CommitteeNodes),
			fmt.Sprintf("%v", chain.Active),
		}
	}
	log.PrintTable(header, rows)
}
