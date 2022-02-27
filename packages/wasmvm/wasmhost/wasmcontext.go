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
	proc      *WasmProcessor
	results   dict.Dict
	sandbox   ISandbox
	wcSandbox *WasmContextSandbox
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

func NewWasmContextForSoloContext(function string, sandbox ISandbox) *WasmContext {
	return &WasmContext{
		funcName:  function,
		sandbox:   sandbox,
		funcTable: NewWasmFuncTable(),
	}
}

func (wc *WasmContext) Call(ctx interface{}) (dict.Dict, error) {
	if wc.id == 0 {
		panic("Context id is zero")
	}

	wc.wcSandbox = NewWasmContextSandbox(wc, ctx)
	wc.sandbox = wc.wcSandbox

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

	wc.trace("Calling " + wc.funcName)
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
		if HostTracing {
			wc.trace("ExportName(%d, %s)", index, name)
		}
		wc.funcTable.SetExport(index, name)
		return
	}

	// index -1 means log WASM tag
	if wc.proc != nil {
		// Invocation through WasmGoVM
		wc.proc.log.Infof("WASM::GO::DEBUG")
		return
	}

	// Invocation through SoloContext
	wc.sandbox.Call(wasmlib.FnLog, wasmtypes.StringToBytes("WASM::SOLO"))
}

func (wc *WasmContext) FunctionFromCode(code uint32) string {
	return wc.funcTable.FunctionFromCode(code)
}

func (wc *WasmContext) IsView() bool {
	return wc.proc.IsView(wc.funcName)
}

func (wc *WasmContext) log() iscp.LogInterface {
	if wc.wcSandbox != nil && wc.wcSandbox.common != nil {
		return wc.wcSandbox.common.Log()
	}
	return wc.proc.log
}

func (wc *WasmContext) Sandbox(funcNr int32, params []byte) []byte {
	if HostTracing && funcNr != wasmlib.FnLog {
		wc.trace("Sandbox(%s)", traceSandbox(funcNr, params))
	}
	res := wc.sandbox.Call(funcNr, params)
	if HostTracing && funcNr != wasmlib.FnLog {
		wc.trace("  => %s", hex(res))
	}
	return res
}

// state reduces the context state to a KVStoreReader
func (wc *WasmContext) state() kv.KVStoreReader {
	ctx := wc.wcSandbox.ctx
	if ctx != nil {
		return ctx.State()
	}
	ctxView := wc.wcSandbox.ctxView
	if ctxView != nil {
		return ctxView.State()
	}
	panic("cannot access state")
}

func (wc *WasmContext) StateDelete(key []byte) {
	if HostTracing {
		wc.trace("StateDelete(%s)", traceHex(key))
	}
	ctx := wc.wcSandbox.ctx
	if ctx == nil {
		panic("StateDelete: readonly state")
	}
	ctx.State().Del(kv.Key(key))
}

func (wc *WasmContext) StateExists(key []byte) bool {
	exists, err := wc.state().Has(kv.Key(key))
	if err != nil {
		panic("StateExists: " + err.Error())
	}
	if HostTracing {
		wc.trace("StateExists(%s) = %v", traceHex(key), exists)
	}
	return exists
}

func (wc *WasmContext) StateGet(key []byte) []byte {
	res, err := wc.state().Get(kv.Key(key))
	if err != nil {
		panic("StateGet: " + err.Error())
	}
	if HostTracing {
		wc.trace("StateGet(%s)", traceHex(key))
		wc.trace("  => %s", hex(res))
	}
	return res
}

func (wc *WasmContext) StateSet(key, value []byte) {
	if HostTracing {
		wc.trace("StateSet(%s, %s)", traceHex(key), traceVal(value))
	}
	ctx := wc.wcSandbox.ctx
	if ctx == nil {
		panic("StateSet: readonly state")
	}
	ctx.State().Set(kv.Key(key), value)
}

func (wc *WasmContext) trace(format string, args ...interface{}) {
	if wc.proc != nil {
		wc.log().Debugf(format, args...)
	}
}

func traceHex(key []byte) string {
	name := ""
	for i, b := range key {
		if b == '.' {
			return string(key[:i+1]) + hex(key[i+1:])
		}
		if b == '#' {
			name = string(key[:i+1])
			j := i + 1
			for ; (key[j] & 0x80) != 0; j++ {
			}
			dec := wasmtypes.NewWasmDecoder(key[i+1 : j+1])
			index := wasmtypes.Uint64Decode(dec)
			name += wasmtypes.Uint64ToString(index)
			if j+1 == len(key) {
				return name
			}
			return name + "..." + hex(key[j+1:])
		}
	}
	return `"` + string(key) + `"`
}

func traceSandbox(funcNr int32, params []byte) string {
	name := sandboxFuncNames[-funcNr]
	if name[0] == '$' {
		return name[1:] + ", " + string(params)
	}
	if name[0] != '#' {
		return name
	}
	return name[1:] + ", " + hex(params)
}

func traceVal(val []byte) string {
	for _, b := range val {
		if b < ' ' || b > '~' {
			return hex(val)
		}
	}
	return string(val)
}

// hex returns a hex string representing the byte buffer
func hex(buf []byte) string {
	const hexa = "0123456789abcdef"
	res := make([]byte, len(buf)*2)
	for i, b := range buf {
		res[i*2] = hexa[b>>4]
		res[i*2+1] = hexa[b&0x0f]
	}
	return string(res)
}
