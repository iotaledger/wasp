// nil processor takes any request and dos nothing, i.e. produces empty state update
// it is useful for testing
package vmnil

import (
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

const ProgramHash = "67F3YgmwXT23PuRwVzDYNLhyXxwQz8WubwmYoWK2hUmE"

type nilProcessor struct {
}

func GetProcessor() vmtypes.Processor {
	return nilProcessor{}
}

func (v nilProcessor) GetEntryPoint(code sctransaction.RequestCode) (vmtypes.EntryPoint, bool) {
	return v, true
}

// does nothing, i.e. resulting state update is empty
func (v nilProcessor) Run(ctx vmtypes.Sandbox) {
	reqId := ctx.AccessRequest().ID()
	ctx.GetWaspLog().Debugw("run nilProcessor",
		"request code", ctx.AccessRequest().Code(),
		"addr", ctx.GetSCAddress().String(),
		"ts", ctx.GetTimestamp(),
		"req", reqId.String(),
	)
}

func (v nilProcessor) WithGasLimit(_ int) vmtypes.EntryPoint {
	return v
}
