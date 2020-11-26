// nil processor takes any request and dos nothing, i.e. produces empty state update
// it is useful for testing
package wasmhost

import (
	"errors"
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

type wasmProcessor struct {
	WasmHost
	ctx       vmtypes.Sandbox
	ctxView   vmtypes.SandboxView
	function  string
	params    codec.ImmutableCodec
	scContext *ScContext
}

func NewWasmProcessor() (*wasmProcessor, error) {
	vm := &wasmProcessor{}
	vm.scContext = NewScContext(vm)
	err := vm.Init(NewNullObject(vm), vm.scContext, &keyMap, vm)
	if err != nil {
		return nil, err
	}
	return vm, nil
}

func (vm *wasmProcessor) call(ctx vmtypes.Sandbox, ctxView vmtypes.SandboxView) (codec.ImmutableCodec, error) {
	saveCtx := vm.ctx
	saveCtxView := vm.ctxView

	vm.ctx = ctx
	vm.ctxView = ctxView
	vm.params = ctx.Params()

	defer func() {
		vm.ctx = saveCtx
		vm.ctxView = saveCtxView
		vm.params = nil
		vm.LogText("Finalizing call")
		vm.scContext.Finalize()
	}()

	testMode, _ := vm.params.Has("testMode")
	if testMode {
		vm.LogText("TEST MODE")
		TestMode = true
	}

	vm.LogText("Calling " + vm.function)
	err := vm.RunScFunction(vm.function)
	if err != nil {
		return nil, err
	}

	if vm.HasError() {
		return nil, errors.New(vm.WasmHost.error)
	}

	resultsId := vm.scContext.GetObjectId(KeyResults, OBJTYPE_MAP)
	results := vm.FindObject(resultsId).(*ScCallResults).Results
	return results, nil
}

func (vm *wasmProcessor) Call(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	return vm.call(ctx, nil)
}

func (vm *wasmProcessor) CallView(ctx vmtypes.SandboxView) (codec.ImmutableCodec, error) {
	return vm.call(nil, ctx)
}

func (vm *wasmProcessor) GetDescription() string {
	return "Wasm VM smart contract processor"
}

func (vm *wasmProcessor) GetEntryPoint(code coretypes.Hname) (vmtypes.EntryPoint, bool) {
	function, ok := vm.codeToFunc[uint32(code)]
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
	vm, err := NewWasmProcessor()
	if err != nil {
		return nil, err
	}
	err = vm.LoadWasm(binaryCode)
	if err != nil {
		return nil, err
	}
	return vm, nil
}

func (vm *wasmProcessor) IsView() bool {
	return (vm.funcToIndex[vm.function] & 0x8000) != 0
}

func (vm *wasmProcessor) SetExport(index int32, functionName string) {
	_, ok := vm.funcToCode[functionName]
	if ok {
		vm.SetError("SetExport: duplicate function name")
		return
	}
	hn := coretypes.Hn(functionName)
	vm.LogText(functionName + " = " + hn.String())
	hashedName := uint32(hn)
	_, ok = vm.codeToFunc[hashedName]
	if ok {
		vm.SetError("SetExport: duplicate hashed name")
		return
	}
	vm.codeToFunc[hashedName] = functionName
	vm.funcToCode[functionName] = hashedName
	vm.funcToIndex[functionName] = index
}

func (vm *wasmProcessor) WithGasLimit(_ int) vmtypes.EntryPoint {
	return vm
}

func (vm *wasmProcessor) Log(logLevel int32, text string) {
	switch logLevel {
	case KeyTraceHost:
		vm.LogText(text)
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
		vm.ctx.Event(text)
		return
	}
	// fallback logging
	fmt.Println(text)
}
