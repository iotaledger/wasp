package wasmproc

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib"
)

const (
	FuncDefault = "_default"
)

type WasmContext struct {
	wasmhost.KvStoreHost
	common   iscp.SandboxBase
	ctx      iscp.Sandbox
	ctxView  iscp.SandboxView
	function string
	id       int32
	host     *wasmhost.WasmHost
	proc     *WasmProcessor
}

var (
	_ iscp.VMProcessorEntryPoint = &WasmContext{}
	_ wasmlib.ScHost             = &WasmContext{}
)

func NewWasmContext(function string, proc *WasmProcessor) *WasmContext {
	wc := &WasmContext{function: function, proc: proc}

	if proc == nil {
		wc.host = &wasmhost.WasmHost{}
		wc.host.Init()
		return wc
	}
	wc.host = &proc.WasmHost

	var scKeys *wasmhost.KvStoreHost
	if proc.scContext != nil {
		scKeys = &proc.scContext.KvStoreHost
	}

	wc.Init(scKeys)
	wc.TrackObject(NewNullObject(&wc.KvStoreHost))
	wc.TrackObject(NewScContext(wc, &wc.KvStoreHost))
	return wc
}

func (wc *WasmContext) AddFunc(f wasmlib.ScFuncContextFunction) []wasmlib.ScFuncContextFunction {
	return wc.host.AddFunc(f)
}

func (wc *WasmContext) AddView(v wasmlib.ScViewContextFunction) []wasmlib.ScViewContextFunction {
	return wc.host.AddView(v)
}

func (wc *WasmContext) Call(ctx interface{}) (dict.Dict, error) {
	if wc.id == 0 {
		panic("Context id is zero")
	}

	wcSaved := wasmhost.Connect(wc)
	defer func() {
		wasmhost.Connect(wcSaved)
		// clean up context after use
		wc.proc.KillContext(wc.id)
	}()

	switch tctx := ctx.(type) {
	case iscp.Sandbox:
		wc.common = tctx
		wc.ctx = tctx
		wc.ctxView = nil
	case iscp.SandboxView:
		wc.common = tctx
		wc.ctx = nil
		wc.ctxView = tctx
	default:
		panic(iscp.ErrWrongTypeEntryPoint)
	}

	if wc.function == "" {
		// init function was missing, do nothing
		return nil, nil
	}

	if wc.function == FuncDefault {
		// TODO default function, do nothing for now
		return nil, nil
	}

	wc.Tracef("Calling " + wc.function)
	err := wc.callFunction()
	if err != nil {
		wc.log().Infof("VM call %s(): error %v", wc.function, err)
		return nil, err
	}
	resultsID := wc.GetObjectID(wasmhost.OBJID_ROOT, wasmhost.KeyResults, wasmhost.OBJTYPE_MAP)
	results := wc.FindObject(resultsID).(*ScDict).kvStore.(dict.Dict)
	return results, nil
}

func (wc *WasmContext) callFunction() error {
	wc.proc.instanceLock.Lock()
	defer wc.proc.instanceLock.Unlock()

	saveID := wc.proc.currentContextID
	wc.proc.currentContextID = wc.id
	err := wc.proc.RunScFunction(wc.function)
	wc.proc.currentContextID = saveID
	return err
}

func (wc *WasmContext) FunctionFromCode(code uint32) string {
	return wc.host.FunctionFromCode(code)
}

func (wc *WasmContext) IsView() bool {
	return wc.host.IsView(wc.function)
}

func (wc *WasmContext) log() iscp.LogInterface {
	if wc.common != nil {
		return wc.common.Log()
	}
	return wc.proc.log
}

func (wc *WasmContext) params() dict.Dict {
	return wc.common.Params()
}

func (wc *WasmContext) state() kv.KVStore {
	if wc.ctx != nil {
		return wc.ctx.State()
	}
	return NewScViewState(wc.ctxView)
}
