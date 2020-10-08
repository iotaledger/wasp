// nil processor takes any request and dos nothing, i.e. produces empty state update
// it is useful for testing
package wasmhost

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

type wasmProcessor struct {
	WasmHost
	ctx vmtypes.Sandbox
}

func GetProcessor(binaryCode []byte) (vmtypes.Processor, error) {
	vm := &wasmProcessor{}
	err := vm.Init(NewNullObject(vm), NewScContext(vm), &keyMap, vm)
	if err != nil {
		return nil, err
	}
	err = vm.LoadWasm(binaryCode)
	if err != nil {
		return nil, err
	}
	return vm, nil
}

func (vm *wasmProcessor) GetEntryPoint(sctransaction.RequestCode) (vmtypes.EntryPoint, bool) {
	//TODO we don't use request code for now, but 'fn' request parameter instead
	return vm, true
}

func (vm *wasmProcessor) GetDescription() string {
	return "Wasm VM smart contract processor"
}

func (vm *wasmProcessor) Run(ctx vmtypes.Sandbox) {
	reqId := ctx.AccessRequest().ID()
	ctx.Publish(fmt.Sprintf("run wasmProcessor: reqCode = %s reqId = %s timestamp = %d",
		ctx.AccessRequest().Code().String(), reqId.String(), ctx.GetTimestamp()))

	functionName, _, _ := ctx.AccessRequest().Args().GetString("fn")
	if functionName == "" {
		ctx.Publish("error starting wasm: Missing fn parameter")
		return
	}

	ctx.Publish("Calling " + functionName)
	vm.ctx = ctx
	err := vm.RunWasmFunction(functionName)
	if err != nil {
		ctx.Publish("error running wasm: " + err.Error())
		panic(err)
	}

	if vm.HasError() {
		errorMsg := vm.GetString(1, KeyError)
		ctx.Publish("error running wasm function: " + errorMsg)
		panic(errorMsg)
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
