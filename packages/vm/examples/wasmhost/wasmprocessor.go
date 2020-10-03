// nil processor takes any request and dos nothing, i.e. produces empty state update
// it is useful for testing
package wasmhost

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

const ProgramHash = "BDREf2rz36AvboHYWfWXgEUG5K8iynLDZAZwKnPBmKM9"

type wasmProcessor struct {
	WasmHost
	ctx vmtypes.Sandbox
}

func GetProcessor() vmtypes.Processor {
	return &wasmProcessor{}
}

func (vm *wasmProcessor) GetEntryPoint(sctransaction.RequestCode) (vmtypes.EntryPoint, bool) {
	// we don't use request code but fn name request parameter instead
	return vm, true
}

func (vm *wasmProcessor) GetDescription() string {
	return "Wasm VM PoC smart contract processor"
}

func (vm *wasmProcessor) Run(ctx vmtypes.Sandbox) {
	reqId := ctx.AccessRequest().ID()
	ctx.Publish(fmt.Sprintf("run wasmProcessor: reqCode = %s reqId = %s timestamp = %d",
		ctx.AccessRequest().Code().String(), reqId.String(), ctx.GetTimestamp()))

	//TODO check what caching optimizations we can do to prevent
	// rebuilding entire object admin and Wasm from scratch on every request
	vm.ctx = ctx
	vm.Init(NewScContext(vm), &keyMap, vm)

	//TODO for now load Wasm code from hardcoded parameter
	// in the future we will need to change things so
	// that we locate the code by hash
	wasm, _, _ := ctx.AccessRequest().Args().GetString("wasm")
	if wasm == "" {
		ctx.Publish("no wasm name specified")
		return
	}
	// when running tests cwd seems to be where the log files go: cluster-data/wasp*
	// use "_bg.wasm" for Rust-based SCs and "_go.wasm" for Go-based SCs
	err := vm.LoadWasm("../../" + wasm + "_go.wasm")
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
}

func (vm *wasmProcessor) WithGasLimit(_ int) vmtypes.EntryPoint {
	return vm
}

func (vm *wasmProcessor) Log(logLevel int32, text string) {
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
