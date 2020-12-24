// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"fmt"
	"github.com/iotaledger/hive.go/logger"
	"strconv"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

type wasmProcessor struct {
	WasmHost
	ctx       vmtypes.Sandbox
	ctxView   vmtypes.SandboxView
	function  string
	nesting   int
	scContext *ScContext
	logger    *logger.Logger
}

func NewWasmProcessor(vm WasmVM, logger *logger.Logger) (*wasmProcessor, error) {
	host := &wasmProcessor{}
	if logger != nil {
		host.logger = logger.Named("wasmtrace")
	}
	host.vm = vm
	host.scContext = NewScContext(host)
	host.Init(NewNullObject(host), host.scContext, host)
	err := host.InitVM(vm)
	if err != nil {
		return nil, err
	}
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
			host.logTextIntern("Finalizing calls")
			host.scContext.Finalize()
		}
		host.ctx = saveCtx
		host.ctxView = saveCtxView
	}()

	testMode, _ := host.Params().Has("testMode")
	if testMode {
		host.logTextIntern("TEST MODE")
		TestMode = true
	}

	host.logTextIntern("Calling " + host.function)
	err := host.RunScFunction(host.function)
	if err != nil {
		return nil, err
	}

	results := host.FindSubObject(nil, KeyResults, OBJTYPE_MAP).(*ScDict).kvStore.(dict.Dict)
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
	function, ok := host.codeToFunc[uint32(code)]
	if !ok && code != coretypes.EntryPointInit {
		return nil, false
	}
	host.function = function
	return host, true
}

func GetProcessor(binaryCode []byte, logger *logger.Logger) (vmtypes.Processor, error) {
	vm, err := NewWasmProcessor(NewWasmTimeVM(), logger)
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
	return (host.funcToIndex[host.function] & 0x8000) != 0
}

func (host *wasmProcessor) SetExport(index int32, functionName string) {
	if index < 0 {
		host.logTextIntern(functionName + " = " + strconv.Itoa(int(index)))
		if index != KeyZzzzzzz {
			host.SetError("SetExport: predefined key value mismatch")
		}
		return
	}
	_, ok := host.funcToCode[functionName]
	if ok {
		host.SetError("SetExport: duplicate function name")
		return
	}
	hn := coretypes.Hn(functionName)
	host.logTextIntern(functionName + " = " + hn.String())
	hashedName := uint32(hn)
	_, ok = host.codeToFunc[hashedName]
	if ok {
		host.SetError("SetExport: duplicate hashed name")
		return
	}
	host.codeToFunc[hashedName] = functionName
	host.funcToCode[functionName] = hashedName
	host.funcToIndex[functionName] = index
}

func (host *wasmProcessor) WithGasLimit(_ int) vmtypes.EntryPoint {
	return host
}

func (host *wasmProcessor) Log(logLevel int32, text string) {
	switch logLevel {
	case KeyTraceAll:
		//host.logTextIntern(text)
	case KeyTrace:
		host.logTextIntern(text)
	case KeyLog:
		host.logTextIntern(text)
	case KeyPanic:
		host.logTextIntern(text)
	case KeyWarning:
		host.logTextIntern(text)
	case KeyError:
		host.logTextIntern(text)
	}
}

// logTextIntern internal tracing for wasmProcessor
func (host *wasmProcessor) logTextIntern(text string) {
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
