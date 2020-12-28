// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"fmt"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
)

// TODO it may be better for the wasmhost to implement Processor interface and
//  use sandbox as call context for some wasmhost call.
//  That would be cleaner architecture: less OO more functional.
//  Alternatively, logger can be passed into the wasm host, because wasmhost needs it for internal tracing

type wasmProcessor struct {
	wasmhost.WasmHost
	ctx       vmtypes.Sandbox
	ctxView   vmtypes.SandboxView
	function  string
	nesting   int
	scContext *ScContext
	logger    *logger.Logger // for internal tracing only
}

var GoWasmVM wasmhost.WasmVM

// NewWasmProcessor creates new wasm processor.
// TODO The logger is for processor tracing, not for sandbox calls
func NewWasmProcessor(vm wasmhost.WasmVM, logger *logger.Logger) (*wasmProcessor, error) {
	host := &wasmProcessor{}
	if logger != nil {
		host.logger = logger.Named("wasmtrace")
	}
	if GoWasmVM != nil {
		vm = GoWasmVM
	}
	err := host.InitVM(vm, false)
	if err != nil {
		return nil, err
	}
	host.scContext = NewScContext(host)
	host.Init(NewNullObject(host), host.scContext, host)
	return host, nil
}

func (host *wasmProcessor) call(ctx vmtypes.Sandbox, ctxView vmtypes.SandboxView) (dict.Dict, error) {
	if host.function == "" {
		// init function was missing, do nothing
		return dict.New(), nil
	}

	saveCtx := host.ctx
	saveCtxView := host.ctxView

	host.ctx = ctx
	host.ctxView = ctxView
	host.nesting++

	defer func() {
		host.nesting--
		if host.nesting == 0 {
			host.logText("Finalizing calls")
			host.scContext.Finalize()
		}
		host.ctx = saveCtx
		host.ctxView = saveCtxView
	}()

	testMode, _ := host.Params().Has("testMode")
	if testMode {
		host.logText("TEST MODE")
		TestMode = true
	}

	host.logText("Calling " + host.function)
	err := host.RunScFunction(host.function)
	if err != nil {
		return nil, err
	}

	results := host.FindSubObject(nil, wasmhost.KeyResults, wasmhost.OBJTYPE_MAP).(*ScDict).kvStore.(dict.Dict)
	return results, nil
}

func (host *wasmProcessor) Call(ctx vmtypes.Sandbox) (dict.Dict, error) {
	return host.call(ctx, nil)
}

func (host *wasmProcessor) CallView(ctx vmtypes.SandboxView) (dict.Dict, error) {
	return host.call(nil, ctx)
}

func (host *wasmProcessor) GetDescription() string {
	return "Wasm VM smart contract processor"
}

func (host *wasmProcessor) GetEntryPoint(code coretypes.Hname) (vmtypes.EntryPoint, bool) {
	function := host.FunctionFromCode(uint32(code))
	if function == "" && code != coretypes.EntryPointInit {
		return nil, false
	}
	host.function = function
	return host, true
}

func GetProcessor(binaryCode []byte, logger *logger.Logger) (vmtypes.Processor, error) {
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

func (host *wasmProcessor) IsView() bool {
	return host.WasmHost.IsView(host.function)
}

func (host *wasmProcessor) WithGasLimit(_ int) vmtypes.EntryPoint {
	return host
}

// TODO in this mixed two different things: internal tracing and Sandbox logging
//  it is not correct semantically
// TODO #2 why do we need so many levels in the SC logging. Logging is mostly for debugging only
//  In the sandbox only Info and Debug (tracing) levels implemented on purpose. Plus special for panic logging
func (host *wasmProcessor) Log(logLevel int32, text string) {
	switch logLevel {
	case wasmhost.KeyTraceAll:
		//host.logText(text)
	case wasmhost.KeyTrace:
		host.logText(text)
	case wasmhost.KeyLog:
		host.logText(text)
	case wasmhost.KeyPanic:
		host.logText(text)
	}
}

// logText internal tracing for wasmProcessor
func (host *wasmProcessor) logText(text string) {
	if host.logger != nil {
		host.logger.Debug(text)
		return
	}
	fmt.Println(text)
}

func (host *wasmProcessor) Balances() coretypes.ColoredBalances {
	if host.ctx != nil {
		return host.ctx.Balances()
	}
	return host.ctxView.Balances()
}

func (host *wasmProcessor) ContractID() coretypes.ContractID {
	if host.ctx != nil {
		return host.ctx.ContractID()
	}
	return host.ctxView.ContractID()
}

func (host *wasmProcessor) Params() dict.Dict {
	if host.ctx != nil {
		return host.ctx.Params()
	}
	return host.ctxView.Params()
}

func (host *wasmProcessor) State() kv.KVStore {
	if host.ctx != nil {
		return host.ctx.State()
	}
	return host.ctxView.State()
}
