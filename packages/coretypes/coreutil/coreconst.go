package coreutil

import (
	"github.com/iotaledger/wasp/packages/coretypes"
)

// names of core contracts
const (
	CoreContractDefault  = "_default"
	CoreContractRoot     = "root"
	CoreContractAccounts = "accounts"
	CoreContractBlob     = "blob"
	CoreContractEventlog = "eventlog"
	CoreContractBlocklog = "blocklog"
)

var (
	CoreContractDefaultHname  = coretypes.Hname(0)
	CoreContractRootHname     = coretypes.Hn(CoreContractRoot)
	CoreContractAccountsHname = coretypes.Hn(CoreContractAccounts)
	CoreContractBlobHname     = coretypes.Hn(CoreContractBlob)
	CoreContractEventlogHname = coretypes.Hn(CoreContractEventlog)
	CoreContractBlocklogHname = coretypes.Hn(CoreContractBlocklog)

	hnames = map[string]coretypes.Hname{
		CoreContractDefault:  CoreContractDefaultHname,
		CoreContractRoot:     CoreContractRootHname,
		CoreContractAccounts: CoreContractAccountsHname,
		CoreContractBlob:     CoreContractBlobHname,
		CoreContractEventlog: CoreContractEventlogHname,
		CoreContractBlocklog: CoreContractBlocklogHname,
	}
)

// the global names used in 'blocklog' contract and in 'state' package
const (
	StateVarTimestamp  = "T"
	StateVarBlockIndex = "I"
)

// used in 'state' package as key for timestamp and block index
var (
	StatePrefixTimestamp  = string(CoreContractBlocklogHname.Bytes()) + StateVarTimestamp
	StatePrefixBlockIndex = string(CoreContractBlocklogHname.Bytes()) + StateVarBlockIndex
)

func CoreHname(name string) coretypes.Hname {
	if ret, ok := hnames[name]; ok {
		return ret
	}
	return coretypes.Hn(name)
}
