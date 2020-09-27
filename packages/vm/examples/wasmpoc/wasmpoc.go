// nil processor takes any request and dos nothing, i.e. produces empty state update
// it is useful for testing
package wasmpoc

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"github.com/iotaledger/wasplib/host"
	"github.com/iotaledger/wasplib/host/interfaces"
	"github.com/iotaledger/wasplib/host/interfaces/level"
	"github.com/iotaledger/wasplib/host/interfaces/objtype"
)

const ProgramHash = "BDREf2rz36AvboHYWfWXgEUG5K8iynLDZAZwKnPBmKM9"
const WasmFolder = "D:/Work/Go/src/github.com/iotaledger/wasplib/wasm/"

type wasmVMPocProcessor struct {
	host.HostBase
	ctx        vmtypes.Sandbox
	entrypoint string
}

func GetProcessor() vmtypes.Processor {
	return &wasmVMPocProcessor{}
}

func (h *wasmVMPocProcessor) GetEntryPoint(sctransaction.RequestCode) (vmtypes.EntryPoint, bool) {
	// we don't use request code but fn name request parameter
	return h, true
}

func (v *wasmVMPocProcessor) GetDescription() string {
	return "Wasm VM PoC smart contract processor"
}

func (h *wasmVMPocProcessor) Run(ctx vmtypes.Sandbox) {
	reqId := ctx.AccessRequest().ID()
	ctx.Publish(fmt.Sprintf("run wasmVMPocProcessor: reqCode = %s reqId = %s timestamp = %d",
		ctx.AccessRequest().Code().String(), reqId.String(), ctx.GetTimestamp()))

	//TODO check what caching optimizations we can do to prevent
	// rebuilding entire object admin and Wasm from scratch on every request
	h.ctx = ctx
	h.Init(h, NewRootObject(h), &keyMap)

	//TODO for now load Wasm code from hardcoded location
	// in the future we will need to change things so
	// that we locate the code by hash and entrypoint
	// by name instead of request code number
	err := h.LoadWasm(WasmFolder + "fairroulette_go.wasm")
	if err != nil {
		ctx.Publish("error loading wasm: " + err.Error())
		return
	}

	functionName, _, _ := ctx.AccessRequest().Args().GetString("fn")
	if functionName == "" {
		ctx.Publish("error starting wasm: Missing fn parameter")
		return
	}
	ctx.Publish("Calling " + functionName)
	err = h.RunWasmFunction(functionName)
	if err != nil {
		ctx.Publish("error running wasm: " + err.Error())
		return
	}

	if h.HasError() {
		ctx.Publish("error running wasm function: " + h.GetString(1, interfaces.KeyError))
		return
	}

	ctx.Publish("Processing transfers...")
	transfersId := h.GetObjectId(1, KeyTransfers, objtype.OBJTYPE_MAP_ARRAY)
	transfers := h.GetObject(transfersId).(*TransfersArray)
	for i := int32(0); i < transfers.GetLength(); i++ {
		transferId := h.GetObjectId(transfersId, i, objtype.OBJTYPE_MAP)
		transfer := h.GetObject(transferId).(*TransferMap)
		transfer.Send()
	}

	ctx.Publish("Processing events...")
	eventsId := h.GetObjectId(1, KeyEvents, objtype.OBJTYPE_MAP_ARRAY)
	events := h.GetObject(eventsId).(*EventsArray)
	for i := int32(0); i < events.GetLength(); i++ {
		requestId := h.GetObjectId(eventsId, i, objtype.OBJTYPE_MAP)
		request := h.GetObject(requestId).(*EventMap)
		request.Send()
	}
}

func (h *wasmVMPocProcessor) WithGasLimit(_ int) vmtypes.EntryPoint {
	return h
}

func (h *wasmVMPocProcessor) Log(logLevel int, text string) {
	if logLevel >= level.TRACE {
		h.ctx.Publish(text)
	}
}
