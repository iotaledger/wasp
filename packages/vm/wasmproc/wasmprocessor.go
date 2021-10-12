// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
)

type WasmProcessor struct {
	wasmhost.WasmHost
	ctx       iscp.Sandbox
	ctxView   iscp.SandboxView
	function  string
	nesting   int
	scContext *ScContext
}

var _ iscp.VMProcessor = &WasmProcessor{}

const (
	FuncDefault      = "_default"
	ViewCopyAllState = "copy_all_state"
)

var GoWasmVM wasmhost.WasmVM

// NewWasmProcessor creates new wasm processor.
func NewWasmProcessor(vm wasmhost.WasmVM, log *logger.Logger) (*WasmProcessor, error) {
	host := &WasmProcessor{}
	if GoWasmVM != nil {
		vm = GoWasmVM
		GoWasmVM = nil
	}
	err := host.InitVM(vm)
	if err != nil {
		return nil, err
	}
	host.scContext = NewScContext(host, &host.KvStoreHost)
	host.Init(log)
	host.TrackObject(NewNullObject(&host.KvStoreHost))
	host.TrackObject(host.scContext)
	host.SetExport(0x8fff, ViewCopyAllState)
	return host, nil
}

func GetProcessor(binaryCode []byte, log *logger.Logger) (iscp.VMProcessor, error) {
	vm, err := NewWasmProcessor(wasmhost.NewWasmTimeVM(), log)
	if err != nil {
		return nil, err
	}
	err = vm.LoadWasm(binaryCode)
	if err != nil {
		return nil, err
	}
	return vm, nil
}

func (host *WasmProcessor) call(ctx iscp.Sandbox, ctxView iscp.SandboxView) (dict.Dict, error) {
	if host.function == "" {
		// init function was missing, do nothing
		return nil, nil
	}

	if host.function == FuncDefault {
		// TODO default function, do nothing for now
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
			host.Tracef("Finalizing calls")
			host.scContext.objects = make(map[int32]int32)
			host.PushFrame()
		}
		host.ctx = saveCtx
		host.ctxView = saveCtxView
	}()

	host.Tracef("Calling " + host.function)
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

func (host *WasmProcessor) Call(ctx interface{}) (dict.Dict, error) {
	switch tctx := ctx.(type) {
	case iscp.Sandbox:
		return host.call(tctx, nil)
	case iscp.SandboxView:
		return host.call(nil, tctx)
	}
	panic(iscp.ErrWrongTypeEntryPoint)
}

func (host *WasmProcessor) GetDescription() string {
	return "Wasm VM smart contract processor"
}

func (host *WasmProcessor) GetEntryPoint(code iscp.Hname) (iscp.VMProcessorEntryPoint, bool) {
	function := host.FunctionFromCode(uint32(code))
	if function == "" && code != iscp.EntryPointInit {
		return nil, false
	}
	host.function = function
	return host, true
}

func (host *WasmProcessor) GetDefaultEntryPoint() iscp.VMProcessorEntryPoint {
	host.function = FuncDefault
	return host
}

func (host *WasmProcessor) IsView() bool {
	return host.WasmHost.IsView(host.function)
}

func (host *WasmProcessor) log() iscp.LogInterface {
	if host.ctx != nil {
		return host.ctx.Log()
	}
	return host.ctxView.Log()
}

func (host *WasmProcessor) params() dict.Dict {
	if host.ctx != nil {
		return host.ctx.Params()
	}
	return host.ctxView.Params()
}

func (host *WasmProcessor) state() kv.KVStore {
	if host.ctx != nil {
		return host.ctx.State()
	}
	return NewScViewState(host.ctxView)
}
