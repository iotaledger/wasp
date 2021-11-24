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

func Get(chainID *iscp.ChainID) *iscp.AgentID {
	return iscp.NewAgentID(chainID.AsAddress(), 0)
}
