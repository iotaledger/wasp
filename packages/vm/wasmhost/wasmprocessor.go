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
	codeToFunc map[int32]string
	ctx        vmtypes.Sandbox
	function   string
	funcToCode map[string]int32
	scContext  *ScContext
}

func GetProcessor(binaryCode []byte) (vmtypes.Processor, error) {
	vm := &wasmProcessor{}
	vm.codeToFunc = make(map[int32]string)
	vm.funcToCode = make(map[string]int32)
	vm.scContext = NewScContext(vm)
	err := vm.Init(NewNullObject(vm), vm.scContext, &keyMap, vm)
	if err != nil {
		return nil, err
	}
	err = vm.LoadWasm(binaryCode)
	if err != nil {
		return nil, err
	}
	err = vm.RunWasmFunction("onLoad")
	if err != nil {
		return nil, err
	}
	return vm, nil
}

func (vm *wasmProcessor) GetEntryPoint(code sctransaction.RequestCode) (vmtypes.EntryPoint, bool) {
	function, ok := vm.codeToFunc[int32(code)]
	if !ok {
		return nil, false
	}
	vm.function = function
	return vm, true
}

func (vm *wasmProcessor) GetDescription() string {
	return "Wasm VM smart contract processor"
}

func (vm *wasmProcessor) Run(ctx vmtypes.Sandbox) {
	reqId := ctx.AccessRequest().ID()
	ctx.Publish(fmt.Sprintf("run wasmProcessor: reqCode = %s reqId = %s timestamp = %d",
		ctx.AccessRequest().Code().String(), reqId.String(), ctx.GetTimestamp()))

	ctx.Publish("Calling " + vm.function)
	vm.ctx = ctx
	err := vm.RunWasmFunction(vm.function)
	if err != nil {
		ctx.Publish("error running wasm: " + err.Error())
		panic(err)
	}

	if vm.HasError() {
		errorMsg := vm.GetString(1, KeyError)
		ctx.Publish("error running wasm function: " + errorMsg)
		panic(errorMsg)
	}

	ctx.Publish("Finalizing call")
	vm.scContext.Finalize()
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
