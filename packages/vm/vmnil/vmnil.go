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

// does nothing, i.e. resulting state update is empty
func (v nilProcessor) Run(ctx vm.Sandbox) {
	reqId := ctx.GetRequestID()
	ctx.GetLog().Debugw("run nilProcessor",
		"request code", ctx.GetRequestCode(),
		"addr", ctx.GetAddress().String(),
		"ts", ctx.GetTimestamp(),
		"state index", ctx.GetStateIndex(),
		"req", reqId.String(),
	)
}
