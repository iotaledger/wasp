package commonaccount

import (
	"github.com/iotaledger/wasp/packages/iscp"
)

var coreHnames = make(map[iscp.Hname]struct{})

func SetCoreHname(hname iscp.Hname) {
	coreHnames[hname] = struct{}{}
}

func IsCoreHname(hname iscp.Hname) bool {
	_, ret := coreHnames[hname]
	return ret
}

// AdjustIfNeeded makes account of the chain owner and all core contracts equal to (chainID, 0)
func AdjustIfNeeded(agentID *iscp.AgentID, chainID *iscp.ChainID) *iscp.AgentID {
	if !agentID.Address().Equals(chainID.AsAddress()) {
		// from another chain
		return agentID
	}
	if IsCoreHname(agentID.Hname()) {
		// one of core contracts
		return Get(chainID)
	}
	return agentID
}

func Get(chainID *iscp.ChainID) *iscp.AgentID {
	return iscp.NewAgentID(chainID.AsAddress(), 0)
}
