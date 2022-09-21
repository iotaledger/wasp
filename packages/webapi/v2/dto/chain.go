package dto

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

type (
	ChainInfo    *governance.ChainInfo
	ContractsMap map[isc.Hname]*root.ContractRecord
)
