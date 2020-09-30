// nil processor takes any request and dos nothing, i.e. produces empty state update
// it is useful for testing
package wasmhost

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

const ProgramHash = "BDREf2rz36AvboHYWfWXgEUG5K8iynLDZAZwKnPBmKM9"
const WasmFolder = "D:/Work/Go/src/github.com/iotaledger/wasplib/wasm/"

type wasmVMPocProcessor struct {
	WasmHost
	ctx vmtypes.Sandbox
}

func GetProcessor() vmtypes.Processor {
	return &wasmVMPocProcessor{}
}

func (vm *wasmVMPocProcessor) GetEntryPoint(sctransaction.RequestCode) (vmtypes.EntryPoint, bool) {
	// we don't use request code but fn name request parameter instead
	return vm, true
}

func (vm *wasmVMPocProcessor) GetDescription() string {
	return "Wasm VM PoC smart contract processor"
}

func (vm *wasmVMPocProcessor) Run(ctx vmtypes.Sandbox) {
	reqId := ctx.AccessRequest().ID()
	ctx.Publish(fmt.Sprintf("run wasmVMPocProcessor: reqCode = %s reqId = %s timestamp = %d",
		ctx.AccessRequest().Code().String(), reqId.String(), ctx.GetTimestamp()))

	//TODO check what caching optimizations we can do to prevent
	// rebuilding entire object admin and Wasm from scratch on every request
	vm.ctx = ctx
	vm.Init(NewRootObject(vm), &keyMap, vm)

	//TODO for now load Wasm code from hardcoded location
	// in the future we will need to change things so
	// that we locate the code by hash and entrypoint
	// by name instead of request code number
	err := vm.LoadWasm(WasmFolder + "fairroulette_go.wasm")
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
	err = vm.RunWasmFunction(functionName)
	if err != nil {
		ctx.Publish("error running wasm: " + err.Error())
		return
	}

	if vm.HasError() {
		ctx.Publish("error running wasm function: " + vm.GetString(1, KeyError))
		return
	}

	ctx.Publish("Processing transfers...")
	transfersId := vm.GetObjectId(1, KeyTransfers, OBJTYPE_MAP_ARRAY)
	transfers := vm.GetObject(transfersId).(*TransfersArray)
	for i := int32(0); i < transfers.GetLength(); i++ {
		transferId := vm.GetObjectId(transfersId, i, OBJTYPE_MAP)
		transfer := vm.GetObject(transferId).(*TransferMap)
		transfer.Send()
	}

	ctx.Publish("Processing events...")
	eventsId := vm.GetObjectId(1, KeyEvents, OBJTYPE_MAP_ARRAY)
	events := vm.GetObject(eventsId).(*EventsArray)
	for i := int32(0); i < events.GetLength(); i++ {
		requestId := vm.GetObjectId(eventsId, i, OBJTYPE_MAP)
		request := vm.GetObject(requestId).(*EventMap)
		request.Send()
	}
}

func (vm *wasmVMPocProcessor) WithGasLimit(_ int) vmtypes.EntryPoint {
	return vm
}

func (vm *wasmVMPocProcessor) Log(logLevel int32, text string) {
	switch logLevel {
	case KeyTraceHost:
		//proc.ctx.Publish(text)
	case KeyTrace:
		vm.ctx.Publish(text)
	case KeyLog:
		vm.ctx.Publish(text)
	case KeyWarning:
		vm.ctx.Publish(text)
	case KeyError:
		vm.ctx.Publish(text)
	}
}
