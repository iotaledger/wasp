// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"sync"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/isc"
)

type WasmProcessor struct {
	contextLock      sync.Mutex
	contexts         map[int32]*WasmContext
	currentContextID int32
	funcTable        *WasmFuncTable
	gasFactorX       uint64
	instanceLock     sync.Mutex
	log              *logger.Logger
	nextContextID    int32
	vm               WasmVM
}

var _ isc.VMProcessor = new(WasmProcessor)

var GoWasmVM func() WasmVM

// GetProcessor creates a new Wasm VM processor.
func GetProcessor(wasmBytes []byte, log *logger.Logger) (isc.VMProcessor, error) {
	proc := &WasmProcessor{
		contexts:   make(map[int32]*WasmContext),
		funcTable:  NewWasmFuncTable(),
		gasFactorX: 1,
		log:        log,
	}

	// By default, we will use WasmTimeVM, but this can be overruled by setting GoWasmVm
	// This setting will also be propagated to all the sub-processors of this processor
	wasmVM := NewWasmTimeVM
	if GoWasmVM != nil {
		wasmVM = GoWasmVM
		GoWasmVM = nil
	}
	proc.vm = wasmVM()

	// load wasm code into a VM Module
	err := proc.vm.LoadWasm(wasmBytes)
	if err != nil {
		return nil, err
	}

	// provide the linker with the sandbox interface
	err = proc.vm.LinkHost()
	if err != nil {
		return nil, err
	}

	wc := NewWasmContext(proc, "")
	wc.vm.GasBudget(1_000_000)
	wc.GasDisable(true)
	Connect(wc)
	proc.currentContextID = wc.id
	err = wc.vm.RunFunction("on_load")
	wc.GasDisable(false)
	//burned := wc.vm.GasBurned()
	//_ = burned
	delete(proc.contexts, wc.id)
	if err != nil {
		return nil, err
	}
	return proc, nil
}

func (proc *WasmProcessor) GetCurrentContext() *WasmContext {
	proc.contextLock.Lock()
	defer proc.contextLock.Unlock()

	return proc.contexts[proc.currentContextID]
}

func (proc *WasmProcessor) GetDefaultEntryPoint() isc.VMProcessorEntryPoint {
	return NewWasmContext(proc, FuncDefault)
}

func (proc *WasmProcessor) GetDescription() string {
	return "Wasm VM smart contract processor"
}

func (proc *WasmProcessor) GetEntryPoint(code isc.Hname) (isc.VMProcessorEntryPoint, bool) {
	function := proc.funcTable.FunctionFromCode(uint32(code))
	if function == "" && code != isc.EntryPointInit {
		return nil, false
	}
	return NewWasmContext(proc, function), true
}

func (proc *WasmProcessor) IsView(function string) bool {
	return (proc.funcTable.funcToIndex[function] & 0x8000) != 0
}

func (proc *WasmProcessor) RegisterContext(wc *WasmContext) {
	proc.contextLock.Lock()
	defer proc.contextLock.Unlock()

	proc.nextContextID++
	wc.id = proc.nextContextID
	proc.contexts[wc.id] = wc
}

func (proc *WasmProcessor) UnregisterContext(wc *WasmContext) {
	proc.contextLock.Lock()
	defer proc.contextLock.Unlock()
	delete(proc.contexts, wc.id)
}

func (proc *WasmProcessor) gasFactor() uint64 {
	return proc.gasFactorX
}
