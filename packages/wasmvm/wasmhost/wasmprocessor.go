// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"errors"
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
	mainProcessor    *WasmProcessor
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

	wc := NewWasmContext("", proc)
	Connect(wc)
	proc.contexts[wc.id] = wc

	// instantiate a new Wasm instance
	err = proc.vm.Instantiate(wc)
	if err != nil {
		return nil, err
	}

	proc.vm.GasBudget(1_000_000)
	proc.vm.GasDisable(true)
	err = proc.vm.RunFunction("on_load")
	proc.vm.GasDisable(false)
	burned := proc.vm.GasBurned()
	_ = burned
	if err != nil {
		return nil, err
	}
	return proc, nil
}

func (proc *WasmProcessor) GetContext() *WasmContext {
	mainProcessor := proc.mainProc()
	mainProcessor.contextLock.Lock()
	defer mainProcessor.contextLock.Unlock()

	return mainProcessor.contexts[mainProcessor.currentContextID]
}

func (proc *WasmProcessor) GetDefaultEntryPoint() iscp.VMProcessorEntryPoint {
	return proc.wasmContext(FuncDefault)
}

func (proc *WasmProcessor) GetDescription() string {
	return "Wasm VM smart contract processor"
}

func (proc *WasmProcessor) GetEntryPoint(code iscp.Hname) (iscp.VMProcessorEntryPoint, bool) {
	function := proc.funcTable.FunctionFromCode(uint32(code))
	if function == "" && code != iscp.EntryPointInit {
		return nil, false
	}
	return proc.wasmContext(function), true
}

func (proc *WasmProcessor) getSubProcessor(vmInstance WasmVM) *WasmProcessor {
	processor := &WasmProcessor{
		log:           proc.log,
		mainProcessor: proc,
		vm:            vmInstance,
	}

	wc := NewWasmContext("", processor)
	Connect(wc)
	err := processor.vm.Instantiate(wc)
	if err != nil {
		panic("cannot instantiate: " + err.Error())
	}
	return processor
}

func (proc *WasmProcessor) IsView(function string) bool {
	return (proc.mainProc().funcTable.funcToIndex[function] & 0x8000) != 0
}

func (proc *WasmProcessor) KillContext(id int32) {
	proc.contextLock.Lock()
	defer proc.contextLock.Unlock()
	delete(proc.contexts, id)
}

func (proc *WasmProcessor) RunScFunction(functionName string) (err error) {
	index, ok := proc.mainProc().funcTable.funcToIndex[functionName]
	if !ok {
		return errors.New("unknown SC function name: " + functionName)
	}
	return proc.vm.RunScFunction(index)
}

func (proc *WasmProcessor) gasFactor() uint64 {
	return proc.mainProc().gasFactorX
}

func (proc *WasmProcessor) mainProc() *WasmProcessor {
	if proc.mainProcessor == nil {
		return proc
	}
	return proc.mainProcessor
}

func (proc *WasmProcessor) wasmContext(function string) *WasmContext {
	processor := proc
	vmInstance := proc.vm.NewInstance()
	if vmInstance != nil {
		processor = proc.getSubProcessor(vmInstance)
	}
	wc := NewWasmContext(function, processor)

	proc.contextLock.Lock()
	defer proc.contextLock.Unlock()

	proc.nextContextID++
	wc.id = proc.nextContextID
	proc.contexts[wc.id] = wc
	return wc
}
