package coreutil

import "github.com/iotaledger/wasp/packages/coretypes"

const (
	OriginBlockEssenceHashBase58 = "6ctFtLGpRNT97JH9NKRu4vyb1atd1ruHmgAxjnf12XTQ"
	OriginStateHashBase58        = "EQB6qXKm2aMHkXXUTL9oMatvWeryQ2Ho6dGc3JT6H3tD"
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
	StatePrefixTimestamp  = string(CoreContractBlobHname.Bytes()) + StateVarTimestamp
	StatePrefixBlockIndex = string(CoreContractBlobHname.Bytes()) + StateVarBlockIndex
)

func CoreHname(name string) coretypes.Hname {
	if ret, ok := hnames[name]; ok {
		return ret
	}
	return coretypes.Hn(name)
}
