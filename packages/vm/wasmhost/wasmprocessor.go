// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"fmt"
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
}

func NewWasmProcessor(vm WasmVM) (*wasmProcessor, error) {
	host := &wasmProcessor{}
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
			host.LogText("Finalizing calls")
			host.scContext.Finalize()
		}
		host.ctx = saveCtx
		host.ctxView = saveCtxView
	}()

	testMode, _ := host.Params().Has("testMode")
	if testMode {
		host.LogText("TEST MODE")
		TestMode = true
	}

	host.LogText("Calling " + host.function)
	err := host.RunScFunction(host.function)
	if err != nil {
		return nil, err
	}

	results := host.FindSubObject(nil, KeyResults, OBJTYPE_MAP).(*ScMutableDict).Dict.(dict.Dict)
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

func (host *wasmProcessor) GetKey(keyId int32) kv.Key {
	return kv.Key(host.GetKeyFromId(keyId))
}

func GetProcessor(binaryCode []byte) (vmtypes.Processor, error) {
	vm, err := NewWasmProcessor(NewWasmTimeVM())
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
		host.LogText(functionName + " = " + strconv.Itoa(int(index)))
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
	host.LogText(functionName + " = " + hn.String())
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
	case KeyTraceHost:
		host.LogText(text)
	case KeyTrace:
		host.LogText(text)
	case KeyLog:
		host.LogText(text)
	case KeyWarning:
		host.LogText(text)
	case KeyError:
		host.LogText(text)
	}
}

// TODO there's a need to distinguish between logging and events
// Also, logging has levels
func (host *wasmProcessor) LogText(text string) {
	if host.ctx != nil {
		//host.ctx.Event(text)
		host.ctx.Log().Infof(text)
		return
	}
	if host.ctxView != nil {
		//host.ctxView.Event(text)
		host.ctxView.Log().Infof(text)
		return
	}
	// fallback logging
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
