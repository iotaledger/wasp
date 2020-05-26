package vmnil

import (
	"github.com/iotaledger/wasp/packages/state"
	vm2 "github.com/iotaledger/wasp/packages/vm"
)

type vmnilstruct struct {
}

func New() vm2.Processor {
	return vmnilstruct{}
}

func (v vmnilstruct) Run(inputs *vm2.VMContext) state.StateUpdate {
	panic("implement me")
}
