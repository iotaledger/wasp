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

func (vm *wasmProcessor) GetDescription() string {
	return "Wasm VM smart contract processor"
}

func (vm *wasmProcessor) GetEntryPoint(code coretypes.Hname) (vmtypes.EntryPoint, bool) {
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
	vm, err := NewWasmProcessor()
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

func (vm *wasmProcessor) Call(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	vm.ctx = ctx
	vm.params = ctx.Params()

	testMode, _ := vm.params.Has("testMode")
	if testMode {
		vm.LogText("TEST MODE")
		TestMode = true
	}
	reqId := ctx.AccessRequest().ID()
	vm.LogText(fmt.Sprintf("run wasmProcessor: reqCode = %s reqId = %s timestamp = %d",
		ctx.AccessRequest().EntryPointCode().String(), reqId.String(), ctx.GetTimestamp()))

	vm.LogText("Calling " + vm.function)
	err := vm.RunFunction(vm.function)
	if err != nil {
		return nil, err
	}

	if vm.HasError() {
		return nil, errors.New(vm.WasmHost.error)
	}

	vm.LogText("Finalizing call")
	vm.scContext.Finalize()
	return nil, nil
}

// TODO
func (ep wasmProcessor) IsView() bool {
	return false
}

// TODO
func (ep wasmProcessor) CallView(ctx vmtypes.SandboxView) (codec.ImmutableCodec, error) {
	panic("implement me")
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
		vm.ctx.Event(text)
		return
	}
	// fallback logging
	fmt.Println(text)
}
