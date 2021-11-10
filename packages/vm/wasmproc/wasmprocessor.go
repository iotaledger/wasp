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
	nextContextID    int32
	scContext        *WasmContext
}

var _ iscp.VMProcessor = &WasmProcessor{}

var GoWasmVM wasmhost.WasmVM

// GetProcessor creates a new Wasm VM processor.
func GetProcessor(binaryCode []byte, log *logger.Logger) (iscp.VMProcessor, error) {
	vm := GoWasmVM
	GoWasmVM = nil
	if vm == nil {
		vm = wasmhost.NewWasmTimeVM()
	}

	proc := &WasmProcessor{log: log, contexts: make(map[int32]*WasmContext)}
	err := proc.InitVM(vm, proc)
	if err != nil {
		return nil, err
	}

	proc.Init()
	// TODO decide if we want be able to examine state directly from tests
	// proc.SetExport(0x8fff, ViewCopyAllState)
	proc.scContext = NewWasmContext("", proc)
	wasmhost.Connect(proc.scContext)
	err = proc.LoadWasm(binaryCode)
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

func (proc *WasmProcessor) wasmContext(function string) *WasmContext {
	wc := NewWasmContext(function, proc)

	proc.contextLock.Lock()
	defer proc.contextLock.Unlock()

	proc.nextContextID++
	wc.id = proc.nextContextID
	proc.contexts[wc.id] = wc
	return wc
}

func (proc *WasmProcessor) GetKvStore(id int32) *wasmhost.KvStoreHost {
	if id == 0 {
		id = proc.currentContextID
	}

	if id == 0 {
		return &proc.scContext.KvStoreHost
	}

	proc.contextLock.Lock()
	defer proc.contextLock.Unlock()

	return &proc.contexts[id].KvStoreHost
}

func (proc *WasmProcessor) KillContext(id int32) {
	proc.contextLock.Lock()
	defer proc.contextLock.Unlock()

	delete(proc.contexts, id)
}
