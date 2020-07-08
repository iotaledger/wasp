// nil processor takes any request and dos nothing, i.e. produces empty state update
// it is useful for testing
package wasmpoc

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

const (
	ProgramHash = "BDREf2rz36AvboHYWfWXgEUG5K8iynLDZAZwKnPBmKM9"

	RequestNOP = sctransaction.RequestCode(uint16(1))
)

type wasmVMPocProcessor struct {
}

func GetProcessor() vmtypes.Processor {
	return wasmVMPocProcessor{}
}

// resolves only RequestNOP
func (v wasmVMPocProcessor) GetEntryPoint(code sctransaction.RequestCode) (vmtypes.EntryPoint, bool) {
	if code == RequestNOP {
		return v, true
	}
	return nil, false
}

// does nothing, i.e. resulting state update is empty
func (v wasmVMPocProcessor) Run(ctx vmtypes.Sandbox) {
	reqId := ctx.AccessRequest().ID()
	ctx.Publish(fmt.Sprintf("run wasmVMPocProcessor: reqCode = %s reqId = %s timestamp = %d",
		ctx.AccessRequest().Code().String(), reqId.String(), ctx.GetTimestamp()))
}

func (v wasmVMPocProcessor) WithGasLimit(_ int) vmtypes.EntryPoint {
	return v
}
