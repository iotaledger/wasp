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
	WasmHost
	contextLock      sync.Mutex
	contexts         map[int32]*WasmContext
	currentContextID int32
	instanceLock     sync.Mutex
	log              *logger.Logger
	mainProcessor    *WasmProcessor
	nextContextID    int32
	scContext        *WasmContext
	wasmVM           func() WasmVM
}

var (
	_ iscp.VMProcessor = new(WasmProcessor)
	_ WasmStore        = new(WasmProcessor)
)

var GoWasmVM func() WasmVM

// GetProcessor creates a new Wasm VM processor.
func GetProcessor(wasmBytes []byte, log *logger.Logger) (iscp.VMProcessor, error) {
	proc := &WasmProcessor{log: log, contexts: make(map[int32]*WasmContext), wasmVM: NewWasmTimeVM}
	proc.Init()

	// By default, we will use WasmTimeVM, but this can be overruled by setting GoWasmVm
	// This setting will also be propagated to all the sub-processors of this processor
	if GoWasmVM != nil {
		proc.wasmVM = GoWasmVM
		GoWasmVM = nil
	}

	// Run setup on main processor, because we will be sharing stuff with the sub-processors
	err := proc.InitVM(proc.wasmVM())
	if err != nil {
		return nil, err
	}

	proc.scContext = NewWasmContext("", proc)
	Connect(proc.scContext)
	err = proc.vm.LoadWasm(wasmBytes)
	if err != nil {
		return nil, err
	}
	err = proc.vm.RunFunction("on_load")
	if err != nil {
		return nil, err
	}
	return proc, nil
}

func (proc *WasmProcessor) GetContext(id int32) *WasmContext {
	if id == 0 {
		id = proc.currentContextID
	}

	if id == 0 {
		return proc.scContext
	}

	mainProcessor := proc.MainProc()
	mainProcessor.contextLock.Lock()
	defer mainProcessor.contextLock.Unlock()

	return mainProcessor.contexts[id]
}

func (proc *WasmProcessor) GetDefaultEntryPoint() iscp.VMProcessorEntryPoint {
	return proc.wasmContext(FuncDefault)
}

func (proc *WasmProcessor) GetDescription() string {
	return "Wasm VM smart contract processor"
}

func (proc *WasmProcessor) GetEntryPoint(code iscp.Hname) (iscp.VMProcessorEntryPoint, bool) {
	function := proc.FunctionFromCode(uint32(code))
	if function == "" && code != iscp.EntryPointInit {
		return nil, false
	}
	return proc.wasmContext(function), true
}

func (proc *WasmProcessor) getSubProcessor(vmInstance WasmVM) *WasmProcessor {
	processor := &WasmProcessor{log: proc.log, mainProcessor: proc, wasmVM: proc.wasmVM}
	processor.Init()
	err := processor.InitVM(vmInstance)
	if err != nil {
		panic("Cannot clone processor: " + err.Error())
	}

	processor.scContext = NewWasmContext("", processor)
	Connect(processor.scContext)
	err = processor.vm.Instantiate()
	if err != nil {
		panic("Cannot instantiate: " + err.Error())
	}
	return processor
}

func (proc *WasmProcessor) KillContext(id int32) {
	proc.contextLock.Lock()
	defer proc.contextLock.Unlock()
	delete(proc.contexts, id)
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

func (proc *WasmProcessor) RunScFunction(functionName string) (err error) {
	index, ok := proc.MainProc().funcToIndex[functionName]
	if !ok {
		return errors.New("unknown SC function name: " + functionName)
	}
	return proc.vm.RunScFunction(index)
}

func (proc *WasmProcessor) IsView(function string) bool {
	return (proc.MainProc().funcToIndex[function] & 0x8000) != 0
}

func (proc *WasmProcessor) MainProc() *WasmProcessor {
	if proc.mainProcessor == nil {
		return proc
	}
	return proc.mainProcessor
}

func (proc *WasmProcessor) InitVM(vm WasmVM) error {
	proc.store = proc
	return vm.LinkHost(vm, &proc.WasmHost)
}
