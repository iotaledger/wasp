// nil processor takes any request and dos nothing, i.e. produces empty state update
// it is useful for testing
package sc6

import (
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

const ProgramHash = "8h2RGcbsUgKckh9rZ4VUF75NUfxP4bj1FC66oSF9us6p"

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
	ctx.GetWaspLog().Debugw("run nilProcessor 6",
		"request code", ctx.AccessRequest().Code(),
		"addr", ctx.GetOwnAddress().String(),
		"ts", ctx.GetTimestamp(),
		"req", reqId.String(),
	)
}

func (v nilProcessor) WithGasLimit(_ int) vmtypes.EntryPoint {
	return v
}
