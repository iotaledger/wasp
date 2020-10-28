// nil processor takes any request and dos nothing, i.e. produces empty state update
// it is useful for testing
package sc7

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

const ProgramHash = "GyJRzfyLGJ43NaDbsd7jFkXhcgxFB13SDcKugAn5pijB"

type nilProcessor struct {
}

func GetProcessor() vmtypes.Processor {
	return nilProcessor{}
}

func (v nilProcessor) GetEntryPoint(code coretypes.EntryPointCode) (vmtypes.EntryPoint, bool) {
	return v, true
}

func (v nilProcessor) GetDescription() string {
	return "Empty (nil) hard coded smart contract processor #7"
}

// does nothing, i.e. resulting state update is empty
func (v nilProcessor) Call(ctx vmtypes.Sandbox, params kv.ImmutableCodec) (kv.ImmutableCodec, error) {
	reqId := ctx.AccessRequest().ID()
	ctx.GetWaspLog().Debugw("run nilProcessor 7",
		"request code", ctx.AccessRequest().Code(),
		"addr", ctx.GetContractID().String(),
		"ts", ctx.GetTimestamp(),
		"req", reqId.String(),
	)
	return nil, nil
}

func (v nilProcessor) WithGasLimit(_ int) vmtypes.EntryPoint {
	return v
}
