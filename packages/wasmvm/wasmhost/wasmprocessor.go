// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"sync"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/iscp"
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

var _ iscp.VMProcessor = new(WasmProcessor)

var GoWasmVM func() WasmVM

// GetProcessor creates a new Wasm VM processor.
func GetProcessor(wasmBytes []byte, log *logger.Logger) (iscp.VMProcessor, error) {
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
	proc.currentContextID = wc.id

	wc.vm.GasBudget(1_000_000)
	wc.GasDisable(true)
	Connect(wc)
	err = wc.vm.RunFunction("on_load")
	wc.GasDisable(false)
	//burned := wc.vm.GasBurned()
	//_ = burned
	proc.KillContext(wc.id)
	if err != nil {
		return nil, err
	}
	return proc, nil
}

func (proc *WasmProcessor) GetContext() *WasmContext {
	proc.contextLock.Lock()
	defer proc.contextLock.Unlock()

	return proc.contexts[proc.currentContextID]
}

func (proc *WasmProcessor) GetDefaultEntryPoint() iscp.VMProcessorEntryPoint {
	return NewWasmContext(proc, FuncDefault)
}

func (proc *WasmProcessor) GetDescription() string {
	return "Wasm VM smart contract processor"
}

func (proc *WasmProcessor) GetEntryPoint(code iscp.Hname) (iscp.VMProcessorEntryPoint, bool) {
	function := proc.funcTable.FunctionFromCode(uint32(code))
	if function == "" && code != iscp.EntryPointInit {
		return nil, false
	}
	return NewWasmContext(proc, function), true
}

func (proc *WasmProcessor) IsView(function string) bool {
	return (proc.funcTable.funcToIndex[function] & 0x8000) != 0
}

func (proc *WasmProcessor) KillContext(id int32) {
	proc.contextLock.Lock()
	defer proc.contextLock.Unlock()
	delete(proc.contexts, id)
}

func (proc *WasmProcessor) gasFactor() uint64 {
	return proc.gasFactorX
}
