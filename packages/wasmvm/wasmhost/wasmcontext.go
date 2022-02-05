package wasmhost

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

const (
	FuncDefault = "_default"
)

type ISandbox interface {
	Call(funcNr int32, params []byte) []byte
}

type WasmContext struct {
	funcName  string
	funcTable *WasmFuncTable
	id        int32
	mini      ISandbox
	proc      *WasmProcessor
	results   dict.Dict
	sandbox   *WasmHostSandbox
}

var (
	_ iscp.VMProcessorEntryPoint = &WasmContext{}
	_ wasmlib.ScHost             = &WasmContext{}
)

func NewWasmContext(function string, proc *WasmProcessor) *WasmContext {
	return &WasmContext{
		funcName:  function,
		proc:      proc,
		funcTable: proc.funcTable,
	}
}

// TODO sensible name?
func NewWasmMiniContext(function string, mini ISandbox) *WasmContext {
	return &WasmContext{
		funcName:  function,
		mini:      mini,
		funcTable: NewWasmFuncTable(),
	}
}

func (wc *WasmContext) Call(ctx interface{}) (dict.Dict, error) {
	if wc.id == 0 {
		panic("Context id is zero")
	}

	wc.sandbox = NewWasmHostSandbox(wc, ctx)
	wc.mini = wc.sandbox

	wcSaved := Connect(wc)
	defer func() {
		Connect(wcSaved)
		// clean up context after use
		wc.proc.KillContext(wc.id)
	}()

	if wc.funcName == "" {
		// init function was missing, do nothing
		return nil, nil
	}

	if wc.funcName == FuncDefault {
		// TODO default function, do nothing for now
		return nil, nil
	}

	wc.log().Debugf("Calling " + wc.funcName)
	wc.results = nil
	err := wc.callFunction()
	if err != nil {
		wc.log().Infof("VM call %s(): error %v", wc.funcName, err)
		return nil, err
	}
	return wc.results, nil
}

func (wc *WasmContext) callFunction() error {
	wc.proc.instanceLock.Lock()
	defer wc.proc.instanceLock.Unlock()

	saveID := wc.proc.currentContextID
	wc.proc.currentContextID = wc.id
	err := wc.proc.RunScFunction(wc.funcName)
	wc.proc.currentContextID = saveID
	return err
}

func (wc *WasmContext) ExportName(index int32, name string) {
	if index >= 0 {
		wc.funcTable.SetExport(index, name)
		return
	}

	// log WASM tag
	if wc.proc != nil {
		wc.proc.log.Infof("WASM::GO::DEBUG")
		return
	}
	if wc.mini != nil {
		wc.mini.Call(wasmlib.FnLog, wasmtypes.StringToBytes("WASM::SOLO"))
		return
	}
	// should never get here
	panic("cannot determine wasm tag: " + name)
}

func (wc *WasmContext) FunctionFromCode(code uint32) string {
	return wc.funcTable.FunctionFromCode(code)
}

func (wc *WasmContext) IsView() bool {
	return wc.proc.IsView(wc.funcName)
}

func (wc *WasmContext) log() iscp.LogInterface {
	if wc.sandbox != nil && wc.sandbox.common != nil {
		return wc.sandbox.common.Log()
	}
	return wc.proc.log
}

func (wc *WasmContext) Sandbox(funcNr int32, params []byte) []byte {
	return wc.mini.Call(funcNr, params)
}

// state reduces the context state to a KVStoreReader
func (wc *WasmContext) state() kv.KVStoreReader {
	ctx := wc.sandbox.ctx
	if ctx != nil {
		return ctx.State()
	}
	ctxView := wc.sandbox.ctxView
	if ctxView != nil {
		return ctxView.State()
	}
	panic("cannot access state")
}

func (wc *WasmContext) StateDelete(key []byte) {
	ctx := wc.sandbox.ctx
	if ctx == nil {
		panic("StateDelete: readonly state")
	}
	ctx.State().Del(kv.Key(key))
}

func (wc *WasmContext) StateExists(key []byte) bool {
	// TODO check err?
	exists, _ := wc.state().Has(kv.Key(key))
	return exists
}

func (wc *WasmContext) StateGet(key []byte) []byte {
	// TODO check err?
	res, _ := wc.state().Get(kv.Key(key))
	return res
}

func (wc *WasmContext) StateSet(key, value []byte) {
	ctx := wc.sandbox.ctx
	if ctx == nil {
		panic("StateSet: readonly state")
	}
	ctx.State().Set(kv.Key(key), value)
}
