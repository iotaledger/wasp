// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"errors"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

const (
	FuncDefault = "_default"
)

type ISandbox interface {
	Call(funcNr int32, params []byte) []byte
	Tracef(format string, args ...interface{})
}

type WasmContext struct {
	funcName    string
	funcTable   *WasmFuncTable
	gasBudget   uint64
	gasBurned   uint64
	gasDisabled bool
	id          int32
	proc        *WasmProcessor
	results     dict.Dict
	sandbox     ISandbox
	vm          WasmVM
	wcSandbox   *WasmContextSandbox
}

var (
	_ isc.VMProcessorEntryPoint = &WasmContext{}
	_ wasmlib.ScHost            = &WasmContext{}
)

func NewWasmContext(proc *WasmProcessor, function string) *WasmContext {
	wc := &WasmContext{
		funcName:  function,
		funcTable: proc.funcTable,
		proc:      proc,
		vm:        proc.vm,
	}
	newInstance := proc.vm.NewInstance(wc)
	if newInstance != nil {
		wc.vm = newInstance
	}
	proc.RegisterContext(wc)
	return wc
}

func NewWasmContextForSoloContext(function string, sandbox ISandbox) *WasmContext {
	return &WasmContext{
		funcName:  function,
		sandbox:   sandbox,
		funcTable: NewWasmFuncTable(),
	}
}

func (wc *WasmContext) Call(ctx interface{}) dict.Dict {
	wc.wcSandbox = NewWasmContextSandbox(wc, ctx)
	wc.sandbox = wc.wcSandbox

	wcSaved := Connect(wc)
	defer func() {
		Connect(wcSaved)
		// clean up context after use
		wc.proc.UnregisterContext(wc)
	}()

	if wc.funcName == "" {
		// init function was missing, do nothing
		return nil
	}

	if wc.funcName == FuncDefault {
		// TODO default function, do nothing for now
		return nil
	}

	wc.log().Debugf("Calling " + wc.funcName)
	wc.results = nil
	err := wc.callFunction()
	if err != nil {
		wc.log().Panicf("VM call %s(): error %v", wc.funcName, err)
	}
	return wc.results
}

func (wc *WasmContext) callFunction() error {
	index, ok := wc.funcTable.funcToIndex[wc.funcName]
	if !ok {
		return errors.New("unknown SC function name: " + wc.funcName)
	}

	proc := wc.proc

	// TODO is this really necessary? We should not be able to call in parallel
	proc.instanceLock.Lock()
	defer proc.instanceLock.Unlock()

	saveID := proc.currentContextID
	proc.currentContextID = wc.id
	wc.gasBudget = wc.GasBudget()
	wc.vm.GasBudget(wc.gasBudget * proc.gasFactor())
	err := wc.vm.RunScFunction(index)
	// if err == nil {
	wc.GasBurned(wc.vm.GasBurned() / proc.gasFactor())
	//}
	wc.gasBurned = wc.gasBudget - wc.GasBudget()
	proc.currentContextID = saveID
	wc.log().Debugf("WC ID %2d, GAS BUDGET %10d, BURNED %10d\n", wc.id, wc.gasBudget, wc.gasBurned)
	return err
}

func (wc *WasmContext) ExportName(index int32, name string) {
	if index >= 0 {
		if HostTracing {
			wc.tracef("ExportName(%d, %s)", index, name)
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

// GasBudget is a callback from the VM that asks for the remaining gas budget of the
// Wasp sandbox. The VM will update the gas budget for the Wasm code with this value
// just before returning to the Wasm code.
func (wc *WasmContext) GasBudget() uint64 {
	if wc.wcSandbox != nil {
		return wc.wcSandbox.common.Gas().Budget()
	}
	return 0
}

// GasBurned is a callback from the VM that sets the remaining gas budget.
// It will update the gas budget for the Wasp sandbox with the amount of gas
// burned by the Wasm code thus far just before calling sandbox.
func (wc *WasmContext) GasBurned(burned uint64) {
	if wc.wcSandbox != nil {
		wc.wcSandbox.common.Gas().Burn(gas.BurnCodeWasm1P, burned)
	}
}

func (wc *WasmContext) GasDisable(disable bool) {
	wc.gasDisabled = disable
}

func (wc *WasmContext) IsView() bool {
	return wc.proc.IsView(wc.funcName)
}

func (wc *WasmContext) log() isc.LogInterface {
	if wc.wcSandbox != nil && wc.wcSandbox.common != nil {
		return wc.wcSandbox.common.Log()
	}
	return wc.proc.log
}

func (wc *WasmContext) RunScFunction(functionName string) (err error) {
	index, ok := wc.funcTable.funcToIndex[functionName]
	if !ok {
		return errors.New("unknown SC function name: " + functionName)
	}
	return wc.vm.RunScFunction(index)
}

func (wc *WasmContext) Sandbox(funcNr int32, params []byte) []byte {
	if !HostTracing || funcNr == wasmlib.FnLog || funcNr == wasmlib.FnTrace {
		return wc.sandbox.Call(funcNr, params)
	}

	wc.tracef("Sandbox(%s)", traceSandbox(funcNr, params))
	// TODO fix this. Probably need to connect proper context or smth
	if wc.sandbox == nil {
		panic("nil sandbox")
	}
	res := wc.sandbox.Call(funcNr, params)
	wc.tracef("  => %s", hex(res))
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
		return ctxView.StateR()
	}
	panic("cannot access state")
}

func (wc *WasmContext) StateDelete(key []byte) {
	if HostTracing {
		wc.tracef("StateDelete(%s)", traceHex(key))
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
		wc.tracef("StateExists(%s) = %v", traceHex(key), exists)
	}
	return exists
}

func (wc *WasmContext) StateGet(key []byte) []byte {
	res, err := wc.state().Get(kv.Key(key))
	if err != nil {
		panic("StateGet: " + err.Error())
	}
	if HostTracing {
		wc.tracef("StateGet(%s)", traceHex(key))
		wc.tracef("  => %s", hex(res))
	}
	return res
}

func (wc *WasmContext) StateSet(key, value []byte) {
	if HostTracing {
		wc.tracef("StateSet(%s, %s)", traceHex(key), traceVal(value))
	}
	ctx := wc.wcSandbox.ctx
	if ctx == nil {
		panic("StateSet: readonly state")
	}
	ctx.State().Set(kv.Key(key), value)
}

func (wc *WasmContext) tracef(format string, args ...interface{}) {
	if wc.proc != nil {
		wc.log().Debugf(format, args...)
		return
	}
	wc.sandbox.Tracef(format, args...)
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
