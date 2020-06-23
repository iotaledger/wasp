// nil processor takes any request and dos nothing, i.e. produces empty state update
// it is useful for testing
package vmnil

import (
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/processor"
)

type nilProcessor struct {
}

func New() processor.Processor {
	return nilProcessor{}
}

func (v nilProcessor) GetEntryPoint(code sctransaction.RequestCode) (processor.EntryPoint, bool) {
	return v, true
}

// does nothing, i.e. resulting state update is empty
func (v nilProcessor) Run(ctx processor.Sandbox) {
	reqId := ctx.Request().ID()
	ctx.GetLog().Debugw("run nilProcessor",
		"request code", ctx.Request().Code(),
		"addr", ctx.GetAddress().String(),
		"ts", ctx.GetTimestamp(),
		"state index", ctx.State().Index(),
		"req", reqId.String(),
	)
}
