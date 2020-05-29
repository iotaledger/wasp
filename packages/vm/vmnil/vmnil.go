package vmnil

import (
	"github.com/iotaledger/wasp/packages/vm"
)

type vmnil struct {
}

func New() vm.Processor {
	return vmnil{}
}

func (v vmnil) Run(inp *vm.VMContext) {
	// does nothing, i.e. resulting state update is empty
	inp.Log.Debugw("run vmnil",
		"addr", inp.Address.String(),
		"color", inp.Color.String(),
		"ts", inp.Timestamp,
		"vs index", inp.VariableState.StateIndex(),
		"req", inp.Request.RequestId().String(),
	)
}
