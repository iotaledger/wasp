package dto

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

type ContractsMap map[isc.Hname]*root.ContractRecord
