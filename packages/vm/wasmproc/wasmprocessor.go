// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
)

type wasmProcessor struct {
	wasmhost.WasmHost
	ctx       coretypes.Sandbox
	ctxView   coretypes.SandboxView
	function  string
	nesting   int
	scContext *ScContext
}

var _ coretypes.VMProcessor = &wasmProcessor{}

const FuncDefault = "_default"
const ViewCopyAllState = "copy_all_state"

var GoWasmVM wasmhost.WasmVM

// NewWasmProcessor creates new wasm processor.
func NewWasmProcessor(vm wasmhost.WasmVM, logger *logger.Logger) (*wasmProcessor, error) {
	host := &wasmProcessor{}
	if GoWasmVM != nil {
		vm = GoWasmVM
		GoWasmVM = nil
	}
	err := host.InitVM(vm, false)
	if err != nil {
		return nil, err
	}
	host.scContext = NewScContext(host)
	host.Init(NewNullObject(&host.KvStoreHost), host.scContext, logger)
	host.SetExport(0x8fff, ViewCopyAllState)
	return host, nil
}

func GetProcessor(binaryCode []byte, logger *logger.Logger) (coretypes.VMProcessor, error) {
	vm, err := NewWasmProcessor(wasmhost.NewWasmTimeVM(), logger)
	if err != nil {
		return nil, err
	}
	err = vm.LoadWasm(binaryCode)
	if err != nil {
		return nil, err
	}
	return vm, nil
}

func (host *wasmProcessor) call(ctx coretypes.Sandbox, ctxView coretypes.SandboxView) (dict.Dict, error) {
	if host.function == "" {
		// init function was missing, do nothing
		return nil, nil
	}

	if host.function == FuncDefault {
		//TODO default function, do nothing for now
		return nil, nil
	}

	if host.function == ViewCopyAllState {
		// dump copy of entire state into result
		state := ctxView.State()
		results := dict.New()
		state.MustIterate("", func(key kv.Key, value []byte) bool {
			results.Set(key, value)
			return true
		})
		return results, nil
	}

	saveCtx := host.ctx
	saveCtxView := host.ctxView

	host.ctx = ctx
	host.ctxView = ctxView
	host.nesting++

	defer func() {
		host.nesting--
		if host.nesting == 0 {
			host.Trace("Finalizing calls")
			host.scContext.objects = make(map[int32]int32)
			host.PushFrame()
		}
		host.ctx = saveCtx
		host.ctxView = saveCtxView
	}()

	testMode, _ := host.params().Has("testMode")
	if testMode {
		host.Trace("TEST MODE")
		TestMode = true
	}

	host.Trace("Calling " + host.function)
	frame := host.PushFrame()
	frameObjects := host.scContext.objects
	host.scContext.objects = make(map[int32]int32)
	err := host.RunScFunction(host.function)
	if err != nil {
		return nil, err
	}
	results := host.FindSubObject(nil, wasmhost.KeyResults, wasmhost.OBJTYPE_MAP).(*ScDict).kvStore.(dict.Dict)
	host.scContext.objects = frameObjects
	host.PopFrame(frame)
	return results, nil
}

func (host *wasmProcessor) Call(ctx interface{}) (dict.Dict, error) {
	switch tctx := ctx.(type) {
	case coretypes.Sandbox:
		return host.call(tctx, nil)
	case coretypes.SandboxView:
		return host.call(nil, tctx)
	}
	panic(coretypes.ErrWrongTypeEntryPoint)
}

func (host *wasmProcessor) GetDescription() string {
	return "Wasm VM smart contract processor"
}

func (host *wasmProcessor) GetEntryPoint(code coretypes.Hname) (coretypes.VMProcessorEntryPoint, bool) {
	function := host.FunctionFromCode(uint32(code))
	if function == "" && code != coretypes.EntryPointInit {
		return nil, false
	}
	host.function = function
	return host, true
}

func (host *wasmProcessor) GetDefaultEntryPoint() coretypes.VMProcessorEntryPoint {
	host.function = FuncDefault
	return host
}

func (host *wasmProcessor) IsView() bool {
	return host.WasmHost.IsView(host.function)
}

func (host *wasmProcessor) accountID() *coretypes.AgentID {
	if host.ctx != nil {
		return host.ctx.AccountID()
	}
	return host.ctxView.AccountID()
}

func (host *wasmProcessor) contract() coretypes.Hname {
	if host.ctx != nil {
		return host.ctx.Contract()
	}
	return host.ctxView.Contract()
}

func (host *wasmProcessor) chainID() *coretypes.ChainID {
	if host.ctx != nil {
		return host.ctx.ChainID()
	}
	return host.ctxView.ChainID()
}

func (host *wasmProcessor) chainOwnerID() *coretypes.AgentID {
	if host.ctx != nil {
		return host.ctx.ChainOwnerID()
	}
	return host.ctxView.ChainOwnerID()
}

func (host *wasmProcessor) contractCreator() *coretypes.AgentID {
	if host.ctx != nil {
		return host.ctx.ContractCreator()
	}
	return host.ctxView.ContractCreator()
}

func (host *wasmProcessor) log() coretypes.LogInterface {
	if host.ctx != nil {
		return host.ctx.Log()
	}
	return host.ctxView.Log()
}

func (host *wasmProcessor) params() dict.Dict {
	if host.ctx != nil {
		return host.ctx.Params()
	}
	return host.ctxView.Params()
}

func (host *wasmProcessor) state() kv.KVStore {
	if host.ctx != nil {
		return host.ctx.State()
	}
	return NewScViewState(host.ctxView)
}

func (host *wasmProcessor) utils() coretypes.Utils {
	if host.ctx != nil {
		return host.ctx.Utils()
	}
	return host.ctxView.Utils()
}
