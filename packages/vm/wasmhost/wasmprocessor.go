// nil processor takes any request and dos nothing, i.e. produces empty state update
// it is useful for testing
package wasmhost

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

type wasmProcessor struct {
	WasmHost
	ctx       vmtypes.Sandbox
	function  string
	scContext *ScContext
}

func (vm *wasmProcessor) GetDescription() string {
	return "Wasm VM smart contract processor"
}

func (vm *wasmProcessor) GetEntryPoint(code sctransaction.RequestCode) (vmtypes.EntryPoint, bool) {
	function, ok := vm.codeToFunc[int32(code)]
	if !ok {
		return nil, false
	}
	vm.function = function
	return vm, true
}

func (vm *wasmProcessor) GetKey(keyId int32) kv.Key {
	return kv.Key(vm.WasmHost.GetKey(keyId))
}

func GetProcessor(binaryCode []byte) (vmtypes.Processor, error) {
	vm := &wasmProcessor{}
	vm.scContext = NewScContext(vm)
	err := vm.Init(NewNullObject(vm), vm.scContext, &keyMap, vm)
	if err != nil {
		return nil, err
	}
	err = vm.LoadWasm(binaryCode)
	if err != nil {
		return nil, err
	}
	err = vm.RunFunction("onLoad")
	if err != nil {
		return nil, err
	}
	return vm, nil
}

func (vm *wasmProcessor) Run(ctx vmtypes.Sandbox) {
	vm.ctx = ctx

	reqId := ctx.AccessRequest().ID()
	vm.LogText(fmt.Sprintf("run wasmProcessor: reqCode = %s reqId = %s timestamp = %d",
		ctx.AccessRequest().Code().String(), reqId.String(), ctx.GetTimestamp()))

	vm.LogText("Calling " + vm.function)
	err := vm.RunFunction(vm.function)
	if err != nil {
		vm.LogText("error running wasm: " + err.Error())
		panic(err)
	}

	if vm.HasError() {
		errorMsg := vm.WasmHost.error
		vm.LogText("error running wasm function: " + errorMsg)
		panic(errorMsg)
	}

	vm.LogText("Finalizing call")
	vm.scContext.Finalize()
}

func (vm *wasmProcessor) WithGasLimit(_ int) vmtypes.EntryPoint {
	return vm
}

func (vm *wasmProcessor) Log(logLevel int32, text string) {
	switch logLevel {
	case KeyTraceHost:
		//vm.LogText(text)
	case KeyTrace:
		vm.LogText(text)
	case KeyLog:
		vm.LogText(text)
	case KeyWarning:
		vm.LogText(text)
	case KeyError:
		vm.LogText(text)
	}
}

func (vm *wasmProcessor) LogText(text string) {
	if vm.ctx != nil {
		vm.ctx.Publish(text)
		return
	}
	// fallback logging
	fmt.Println(text)
}
