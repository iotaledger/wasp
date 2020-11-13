package accountsc

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

type accountsProcessor map[coretypes.Hname]accountsEntryPoint

type epFunc func(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error)
type epFuncView func(ctx vmtypes.SandboxView) (codec.ImmutableCodec, error)

type accountsEntryPoint struct {
	fun interface{}
}

const (
	FuncFetchAccount = "fetchAccount"
	ContractName     = "accounts"

	VarStateInitialized = "i"
)

var (
	Hname                  = coretypes.Hn(ContractName)
	EntryPointFetchAccount = coretypes.Hn(FuncFetchAccount)

	processor = accountsProcessor{
		coretypes.EntryPointCodeInit: {epFunc(initialize)},
		EntryPointFetchAccount:       {epFunc(fetchAccount)},
	}
	ProgramHash = hashing.NilHash
)

func GetProcessor() vmtypes.Processor {
	return processor
}

func (v accountsProcessor) GetEntryPoint(code coretypes.Hname) (vmtypes.EntryPoint, bool) {
	ret, ok := processor[code]
	return ret, ok
}

func (v accountsProcessor) GetDescription() string {
	return "Acount processor"
}

func (ep accountsEntryPoint) Call(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	fun, ok := ep.fun.(epFunc)
	if !ok {
		return nil, fmt.Errorf("wrong type of entry point")
	}
	ret, err := fun(ctx)
	if err != nil {
		ctx.Eventf("error occurred: '%v'", err)
	}
	return ret, err
}

func (ep accountsEntryPoint) IsView() bool {
	_, ok := ep.fun.(epFuncView)
	return ok
}

func (ep accountsEntryPoint) CallView(ctx vmtypes.SandboxView) (codec.ImmutableCodec, error) {
	fun, ok := ep.fun.(epFuncView)
	if !ok {
		return nil, fmt.Errorf("wrong type of entry point")
	}
	return fun(ctx)
}

func (ep accountsEntryPoint) WithGasLimit(_ int) vmtypes.EntryPoint {
	return ep
}
