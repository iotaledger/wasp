package chain

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
)

func infoCmd(args []string) {
	chain, err := config.WaspClient().GetChainRecord(GetCurrentChainID())
	check(err)
	showChainInfo(chain)
}

func showChainInfo(chain *registry.ChainRecord) {
	fmt.Printf(
		"%s: [color %s] [committee %s] [active? %v]\n",
		chain.ChainID,
		chain.Color,
		chain.CommitteeNodes,
		chain.Active,
	)
}
