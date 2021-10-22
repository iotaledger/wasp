package wasmproc

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
	"github.com/iotaledger/wasp/packages/vm/wasmlib"
)

const (
	FuncDefault      = "_default"
	ViewCopyAllState = "copy_all_state"
)

type WasmContext struct {
	wasmhost.KvStoreHost
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

func (wc *WasmContext) AddFunc(f func(ctx wasmlib.ScFuncContext)) []func(ctx wasmlib.ScFuncContext) {
	return wc.host.AddFunc(f)
}

func (wc *WasmContext) AddView(v func(ctx wasmlib.ScViewContext)) []func(ctx wasmlib.ScViewContext) {
	return wc.host.AddView(v)
}

func (wc *WasmContext) Call(ctx interface{}) (dict.Dict, error) {
	if wc.id == 0 {
		panic("Context id is zero")
	}

	wcSaved := wasmlib.ConnectHost(wc)
	defer func() {
		wasmlib.ConnectHost(wcSaved)
		// clean up context after use
		wc.proc.KillContext(wc.id)
	}()

	switch tctx := ctx.(type) {
	case iscp.Sandbox:
		wc.ctx = tctx
	case iscp.SandboxView:
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

	// TODO decide if we want be able to examine state directly from tests
	//if wc.function == ViewCopyAllState {
	//	// dump copy of entire state into result
	//	state := wc.ctxView.State()
	//	results := dict.New()
	//	state.MustIterate("", func(key kv.Key, value []byte) bool {
	//		results.Set(key, value)
	//		return true
	//	})
	//	return results, nil
	//}

	wc.Tracef("Calling " + wc.function)
	err := wc.callFunction()
	if err != nil {
		wc.log().Infof("VM call %s(): error %v", wc.function, err)
		return nil, err
	}
	resultsID := wc.GetObjectID(wasmlib.OBJ_ID_ROOT, wasmhost.KeyResults, wasmhost.OBJTYPE_MAP)
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
	if wc.ctx != nil {
		return wc.ctx.Log()
	}
	if wc.ctxView != nil {
		return wc.ctxView.Log()
	}
	return wc.proc.log
}

func (wc *WasmContext) params() dict.Dict {
	if wc.ctx != nil {
		return wc.ctx.Params()
	}
	return wc.ctxView.Params()
}

func (wc *WasmContext) state() kv.KVStore {
	if wc.ctx != nil {
		return wc.ctx.State()
	}
	return NewScViewState(wc.ctxView)
}
