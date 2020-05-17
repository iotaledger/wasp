package state

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/hashing"
)

type NewOriginBatchParams struct {
	Address      *address.Address
	OwnerAddress *address.Address
	Description  string
	ProgramHash  *hashing.HashValue
}

func NewOriginBatch(par NewOriginBatchParams) Batch {
	stateUpd := NewStateUpdate(nil)
	vars := stateUpd.Variables()
	vars.Set("$address$", par.Address.String())
	vars.Set("$owneraddr$", par.OwnerAddress.String())
	vars.Set("$descr$", par.Description)
	vars.Set("$proghash$", par.ProgramHash.String())
	ret, err := NewBatch([]StateUpdate{stateUpd}, 0)
	if err != nil {
		panic(err)
	}
	return ret
}

func OriginVariableStateHash(par NewOriginBatchParams) *hashing.HashValue {
	batch := NewOriginBatch(par)
	originState := NewVariableState(nil)
	if err := originState.Apply(batch); err != nil {
		panic(err)
	}
	return originState.Hash()
}
