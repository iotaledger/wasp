package chain

import (
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func chainCommittee() []int {
	chain, err := config.WaspClient().GetChainRecord(GetCurrentChainID())
	log.Check(err)

	r := []int{}
	for _, peering := range chain.CommitteeNodes {
		r = append(r, config.FindNodeBy(config.HostKindPeering, peering))
	}
	return r
}
