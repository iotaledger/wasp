package coreutil

import (
	"github.com/iotaledger/wasp/packages/isc"
)

// names of core contracts
const (
	CoreContractDefault         = "_default"
	CoreContractRoot            = "root"
	CoreContractAccounts        = "accounts"
	CoreContractBlob            = "blob"
	CoreContractEventlog        = "eventlog"
	CoreContractBlocklog        = "blocklog"
	CoreContractErrors          = "errors"
	CoreContractGovernance      = "governance"
	CoreEPRotateStateController = "rotateStateController"
)

var (
	CoreContractDefaultHname         = isc.Hname(0)
	CoreContractRootHname            = isc.Hn(CoreContractRoot)
	CoreContractAccountsHname        = isc.Hn(CoreContractAccounts)
	CoreContractBlobHname            = isc.Hn(CoreContractBlob)
	CoreContractEventlogHname        = isc.Hn(CoreContractEventlog)
	CoreContractBlocklogHname        = isc.Hn(CoreContractBlocklog)
	CoreContractErrorsHname          = isc.Hn(CoreContractErrors)
	CoreContractGovernanceHname      = isc.Hn(CoreContractGovernance)
	CoreEPRotateStateControllerHname = isc.Hn(CoreEPRotateStateController)

	hnames = map[string]isc.Hname{
		CoreContractDefault:    CoreContractDefaultHname,
		CoreContractRoot:       CoreContractRootHname,
		CoreContractAccounts:   CoreContractAccountsHname,
		CoreContractBlob:       CoreContractBlobHname,
		CoreContractEventlog:   CoreContractEventlogHname,
		CoreContractBlocklog:   CoreContractBlocklogHname,
		CoreContractGovernance: CoreContractGovernanceHname,
		CoreContractErrors:     CoreContractErrorsHname,
	}
)

// the global names used in 'blocklog' contract and in 'state' package
const (
	StateVarTimestamp           = "T"
	StateVarBlockIndex          = "I"
	StateVarPrevL1Commitment    = "H"
	ParamStateControllerAddress = "S"
)

// used in 'state' package as key for timestamp and block index
var (
	StatePrefixTimestamp        = string(CoreContractBlocklogHname.Bytes()) + StateVarTimestamp
	StatePrefixBlockIndex       = string(CoreContractBlocklogHname.Bytes()) + StateVarBlockIndex
	StatePrefixPrevL1Commitment = string(CoreContractBlocklogHname.Bytes()) + StateVarPrevL1Commitment
)

func CoreHname(name string) isc.Hname {
	if ret, ok := hnames[name]; ok {
		return ret
	}
	return isc.Hn(name)
}
