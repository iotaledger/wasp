// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"sync"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
)

type WasmProcessor struct {
	wasmhost.WasmHost
	contextLock      sync.Mutex
	contexts         map[int32]*WasmContext
	currentContextID int32
	instanceLock     sync.Mutex
	log              *logger.Logger
	mainProcessor    *WasmProcessor
	nextContextID    int32
	// procesorPool     *WasmProcessor
	scContext *WasmContext
	wasmBytes []byte
	wasmVM    func() wasmhost.WasmVM
}

var _ iscp.VMProcessor = &WasmProcessor{}

var GoWasmVM func() wasmhost.WasmVM

// GetProcessor creates a new Wasm VM processor.
func GetProcessor(wasmBytes []byte, log *logger.Logger) (iscp.VMProcessor, error) {
	// By default we will use WasmTimeVM, but this can be overruled by setting GoWasmVm
	// This setting will be propagated to all the sub-processors of this processor
	vm := GoWasmVM
	GoWasmVM = nil
	if vm == nil {
		vm = wasmhost.NewWasmTimeVM
	}

	// run setup on main processor, because we will be sharing stuff with the sub-processors
	proc := &WasmProcessor{log: log, contexts: make(map[int32]*WasmContext), wasmBytes: wasmBytes, wasmVM: vm}
	err := proc.InitVM(vm(), proc)
	if err != nil {
		return nil, err
	}

	proc.Init()
	// TODO decide if we want be able to examine state directly from tests
	// proc.SetExport(0x8fff, ViewCopyAllState)
	proc.scContext = NewWasmContext("", proc)
	wasmhost.Connect(proc.scContext)
	err = proc.LoadWasm(proc.wasmBytes)
	if err != nil {
		return nil, err
	}
	return proc, nil
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

func (proc *WasmProcessor) GetDefaultEntryPoint() iscp.VMProcessorEntryPoint {
	return proc.wasmContext(FuncDefault)
}

func (proc *WasmProcessor) GetKvStore(id int32) *wasmhost.KvStoreHost {
	if id == 0 {
		id = proc.currentContextID
	}

	if id == 0 {
		return &proc.scContext.KvStoreHost
	}

	mainProcessor := proc
	if proc.mainProcessor != nil {
		mainProcessor = proc.mainProcessor
	}
	mainProcessor.contextLock.Lock()
	defer mainProcessor.contextLock.Unlock()

	return &mainProcessor.contexts[id].KvStoreHost
}

func (proc *WasmProcessor) KillContext(id int32) {
	proc.contextLock.Lock()
	defer proc.contextLock.Unlock()

	// TODO release processor to pool? In that case, when taking it out, don't forget to reset its data

	// owner := proc.proc
	// proc.proc = owner.pool
	// owner.pool = proc

	delete(proc.contexts, id)
}

func (proc *WasmProcessor) wasmContext(function string) *WasmContext {
	// clone and setup processor for each context
	processor := proc
	if proc.WasmHost.PoolSize() != 0 {
		processor = proc.getProcessor()
	}
	wc := NewWasmContext(function, processor)

	proc.contextLock.Lock()
	defer proc.contextLock.Unlock()

	proc.nextContextID++
	wc.id = proc.nextContextID
	proc.contexts[wc.id] = wc
	return wc
}

func (proc *WasmProcessor) getProcessor() *WasmProcessor {
	//processor = proc.fromPool()
	//if processor != nil {
	//	err := processor.WasmHost.Instantiate()
	//	if err != nil {
	//		panic("Cannot instantiate processor: " + err.Error())
	//	}
	//	return processor
	//}

	processor := &WasmProcessor{log: proc.log, mainProcessor: proc, wasmBytes: proc.wasmBytes, wasmVM: proc.wasmVM}
	err := processor.InitVM(proc.NewInstance(), processor)
	if err != nil {
		panic("Cannot clone processor: " + err.Error())
	}
	processor.Init()
	// TODO decide if we want be able to examine state directly from tests
	// proc.SetExport(0x8fff, ViewCopyAllState)
	processor.scContext = NewWasmContext("", processor)
	wasmhost.Connect(processor.scContext)
	err = processor.Instantiate()
	if err != nil {
		panic("Cannot instantiate: " + err.Error())
	}
	err = processor.RunFunction("on_load")
	if err != nil {
		panic("Cannot run on_load: " + err.Error())
	}
	return processor
}

//func (proc *WasmProcessor) fromPool() *WasmProcessor {
//	proc.contextLock.Lock()
//	defer proc.contextLock.Unlock()
//	processor := proc.procesorPool
//	if processor != nil {
//		proc.procesorPool = processor.mainProcessor
//		processor.mainProcessor = proc
//	}
//	return processor
//}
