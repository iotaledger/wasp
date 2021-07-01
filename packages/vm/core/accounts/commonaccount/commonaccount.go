package commonaccount

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/chainid"
)

var coreHnames = make(map[coretypes.Hname]struct{})

func SetCoreHname(hname coretypes.Hname) {
	coreHnames[hname] = struct{}{}
}

func IsCoreHname(hname coretypes.Hname) bool {
	_, ret := coreHnames[hname]
	return ret
}

// AdjustIfNeeded makes account of the chain owner and all core contracts equal to (chainID, 0)
func AdjustIfNeeded(agentID *coretypes.AgentID, chainID *chainid.ChainID) *coretypes.AgentID {
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

func Get(chainID *chainid.ChainID) *coretypes.AgentID {
	return coretypes.NewAgentID(chainID.AsAddress(), 0)
}
