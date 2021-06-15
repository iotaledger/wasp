package commonaccount

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/chainid"
)

var coreHnames = make(map[coretypes.Hname]struct{})

func SetCoreHname(hname coretypes.Hname) {
	coreHnames[hname] = struct{}{}
}

func isCoreHname(hname coretypes.Hname) bool {
	_, ret := coreHnames[hname]
	return ret
}

// Adjust makes account of the chain owner and all core contracts equal to (chainID, 0)
func Adjust(agentID *coretypes.AgentID, chainID *chainid.ChainID, chainOwnerID *coretypes.AgentID) *coretypes.AgentID {
	if chainOwnerID.Equals(agentID) {
		// chain owner
		return Get(chainID)
	}
	if !agentID.Address().Equals(chainID.AsAddress()) {
		// from another chain
		return agentID
	}
	if isCoreHname(agentID.Hname()) {
		// one of core contracts
		return Get(chainID)
	}
	return agentID
}

func Get(chainID *chainid.ChainID) *coretypes.AgentID {
	return coretypes.NewAgentID(chainID.AsAddress(), 0)
}
