// nil processor takes any request and dos nothing, i.e. produces empty state update
// it is useful for testing
package vmnil

import (
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/processor"
	"github.com/iotaledger/wasp/packages/vm/sandbox"
)

const ProgramHash = "67F3YgmwXT23PuRwVzDYNLhyXxwQz8WubwmYoWK2hUmE"

type nilProcessor struct {
}

func New() processor.Processor {
	return nilProcessor{}
}

func (v nilProcessor) GetEntryPoint(code sctransaction.RequestCode) (processor.EntryPoint, bool) {
	return v, true
}

// does nothing, i.e. resulting state update is empty
func (v nilProcessor) Run(ctx sandbox.Sandbox) {
	reqId := ctx.AccessRequest().ID()
	ctx.GetLog().Debugw("run nilProcessor",
		"request code", ctx.AccessRequest().Code(),
		"addr", ctx.GetAddress().String(),
		"ts", ctx.GetTimestamp(),
		"req", reqId.String(),
	)
}
