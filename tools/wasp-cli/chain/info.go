package chain

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func infoCmd(args []string) {
	chain, err := config.WaspClient().GetChainRecord(GetCurrentChainID())
	log.Check(err)
	showChainInfo([]*registry.ChainRecord{chain})
}

func showChainInfo(chains []*registry.ChainRecord) {
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
