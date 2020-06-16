package vmnil

import (
	"github.com/iotaledger/wasp/packages/sctransaction"
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
		"ts", inp.Timestamp,
		"state index", inp.VariableState.StateIndex(),
		"req", inp.Request.RequestId().String(),
	)
}

// all calls to this processor are equivalent to NOP
func (v vmnil) Supports(code sctransaction.RequestCode) bool {
	return false
}
