package chain

import (
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func chainCommittee() []int {
	chainID := GetCurrentChainID()
	committee, err := config.WaspClient().GetCommitteeForChain(chainID)
	log.Check(err)

	r := []int{}
	for _, peering := range committee.Nodes {
		r = append(r, config.FindNodeBy(config.HostKindPeering, peering))
	}
	return r
}
