package coreutil

import (
	"github.com/iotaledger/wasp/packages/iscp"
)

// names of core contracts
const (
	CoreContractDefault         = "_default"
	CoreContractRoot            = "root"
	CoreContractAccounts        = "accounts"
	CoreContractBlob            = "blob"
	CoreContractEventlog        = "eventlog"
	CoreContractBlocklog        = "blocklog"
	CoreContractError           = "error"
	CoreContractGovernance      = "governance"
	CoreEPRotateStateController = "rotateStateController"
)

var (
	CoreContractDefaultHname         = iscp.Hname(0)
	CoreContractRootHname            = iscp.Hn(CoreContractRoot)
	CoreContractAccountsHname        = iscp.Hn(CoreContractAccounts)
	CoreContractBlobHname            = iscp.Hn(CoreContractBlob)
	CoreContractEventlogHname        = iscp.Hn(CoreContractEventlog)
	CoreContractBlocklogHname        = iscp.Hn(CoreContractBlocklog)
	CoreContractErrorHname           = iscp.Hn(CoreContractError)
	CoreContractGovernanceHname      = iscp.Hn(CoreContractGovernance)
	CoreEPRotateStateControllerHname = iscp.Hn(CoreEPRotateStateController)

	hnames = map[string]iscp.Hname{
		CoreContractDefault:    CoreContractDefaultHname,
		CoreContractRoot:       CoreContractRootHname,
		CoreContractAccounts:   CoreContractAccountsHname,
		CoreContractBlob:       CoreContractBlobHname,
		CoreContractEventlog:   CoreContractEventlogHname,
		CoreContractBlocklog:   CoreContractBlocklogHname,
		CoreContractGovernance: CoreContractGovernanceHname,
		CoreContractError:      CoreContractErrorHname,
	}
)

// the global names used in 'blocklog' contract and in 'state' package
const (
	StateVarTimestamp           = "T"
	StateVarBlockIndex          = "I"
	StateVarPrevStateHash       = "H"
	ParamStateControllerAddress = "S"
)

// used in 'state' package as key for timestamp and block index
var (
	StatePrefixTimestamp     = string(CoreContractBlocklogHname.Bytes()) + StateVarTimestamp
	StatePrefixBlockIndex    = string(CoreContractBlocklogHname.Bytes()) + StateVarBlockIndex
	StatePrefixPrevStateHash = string(CoreContractBlocklogHname.Bytes()) + StateVarPrevStateHash
)

func CoreHname(name string) iscp.Hname {
	if ret, ok := hnames[name]; ok {
		return ret
	}
	return iscp.Hn(name)
}
