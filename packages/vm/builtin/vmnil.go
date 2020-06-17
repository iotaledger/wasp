package builtin

import (
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm"
)

type builtinProcessor struct {
}

type builtinEntryPoint struct {
}

var Processor = New()

func New() vm.Processor {
	return &builtinProcessor{}
}

func (v *builtinProcessor) GetEntryPoint(code sctransaction.RequestCode) (vm.EntryPoint, bool) {
	if !code.IsReserved() {
		return nil, false
	}
	return &builtinEntryPoint{}, true
}

func (v *builtinEntryPoint) Run(inp *vm.VMContext) {
	// does nothing, i.e. resulting state update is empty
	inp.Log.Debugw("run fake builtin processor",
		"addr", inp.Address.String(),
		"ts", inp.Timestamp,
		"state index", inp.VariableState.StateIndex(),
		"req", inp.Request.RequestId().String(),
	)
}
