// nil processor takes any request and dos nothing, i.e. produces empty state update
// it is useful for testing
package vmnil

import (
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm"
)

type nilProcessor struct {
}

func New() vm.Processor {
	return nilProcessor{}
}

func (v nilProcessor) GetEntryPoint(code sctransaction.RequestCode) (vm.EntryPoint, bool) {
	return v, true
}

func (v nilProcessor) Run(inp *vm.VMContext) {
	// does nothing, i.e. resulting state update is empty
	inp.Log.Debugw("run nilprocessor",
		"addr", inp.Address.String(),
		"ts", inp.Timestamp,
		"state index", inp.VariableState.StateIndex(),
		"req", inp.Request.RequestId().String(),
	)
}
