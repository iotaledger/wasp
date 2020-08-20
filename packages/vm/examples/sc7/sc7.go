// nil processor takes any request and dos nothing, i.e. produces empty state update
// it is useful for testing
package sc7

import (
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

const ProgramHash = "GyJRzfyLGJ43NaDbsd7jFkXhcgxFB13SDcKugAn5pijB"

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
	ctx.GetWaspLog().Debugw("run nilProcessor 7",
		"request code", ctx.AccessRequest().Code(),
		"addr", ctx.GetSCAddress().String(),
		"ts", ctx.GetTimestamp(),
		"req", reqId.String(),
	)
}

func (v nilProcessor) WithGasLimit(_ int) vmtypes.EntryPoint {
	return v
}
