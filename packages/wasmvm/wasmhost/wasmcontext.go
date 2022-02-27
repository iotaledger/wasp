// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/iscp"
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
}

type WasmContext struct {
	funcName  string
	funcTable *WasmFuncTable
	gasBudget uint64
	gasBurned uint64
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

func (wc *WasmContext) Call(ctx interface{}) dict.Dict {
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
	wc.proc.instanceLock.Lock()
	defer wc.proc.instanceLock.Unlock()

	saveID := wc.proc.currentContextID
	wc.proc.currentContextID = wc.id
	wc.gasBudget = wc.GasBudget()
	wc.proc.vm.GasBudget(wc.gasBudget * wc.proc.gasFactor())
	err := wc.proc.RunScFunction(wc.funcName)
	if err == nil {
		wc.GasBurned(wc.proc.vm.GasBurned() / wc.proc.gasFactor())
	}
	wc.gasBurned = wc.gasBudget - wc.GasBudget()
	wc.proc.currentContextID = saveID
	fmt.Printf("WC ID %2d, GAS BUDGET %10d, BURNED %10d\n", wc.id, wc.gasBudget, wc.gasBurned)
	return err
}

func (wc *WasmContext) ExportName(index int32, name string) {
	if index >= 0 {
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
	return wc.sandbox.Call(funcNr, params)
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
	return exists
}

func (wc *WasmContext) StateGet(key []byte) []byte {
	res, err := wc.state().Get(kv.Key(key))
	if err != nil {
		panic("StateGet: " + err.Error())
	}
	return res
}

func (wc *WasmContext) StateSet(key, value []byte) {
	ctx := wc.wcSandbox.ctx
	if ctx == nil {
		panic("StateSet: readonly state")
	}
	ctx.State().Set(kv.Key(key), value)
}
