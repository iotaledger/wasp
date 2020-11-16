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
	ContractName = "accounts"

	FuncBalance  = "balance"
	FuncDeposit  = "deposit"
	FuncMove     = "move"
	FuncWithdraw = "withdraw"
	FuncAccounts = "accounts"

	VarStateInitialized = "i"
	VarStateAllAccounts = "a"

	ParamAgentID = "a"
)

var (
	Hname              = coretypes.Hn(ContractName)
	EntryPointBalance  = coretypes.Hn(FuncBalance)
	EntryPointDeposit  = coretypes.Hn(FuncDeposit)
	EntryPointMove     = coretypes.Hn(FuncMove)
	EntryPointWithdraw = coretypes.Hn(FuncWithdraw)
	EntryPointAccounts = coretypes.Hn(FuncAccounts)

	processor = accountsProcessor{
		coretypes.EntryPointInit: {epFunc(initialize)},
		EntryPointBalance:        {epFuncView(getBalance)},
		EntryPointAccounts:       {epFuncView(getAccounts)},
		EntryPointDeposit:        {epFunc(deposit)},
		EntryPointMove:           {epFunc(move)},
		EntryPointWithdraw:       {epFunc(withdraw)},
	}
	ProgramHash = hashing.NilHash

	ErrParamsAgentIDNotFound = fmt.Errorf("wrong parameters: agent ID not specified")
)

func GetProcessor() vmtypes.Processor {
	return processor
}

func (v accountsProcessor) GetEntryPoint(code coretypes.Hname) (vmtypes.EntryPoint, bool) {
	ret, ok := processor[code]
	return ret, ok
}

func (v accountsProcessor) GetDescription() string {
	return "Accounts processor"
}

func (ep accountsEntryPoint) Call(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	fun, ok := ep.fun.(epFunc)
	if !ok {
		return nil, vmtypes.ErrWrongTypeEntryPoint
	}
	ret, err := fun(ctx)
	if err != nil {
		ctx.Eventf("accountsEntryPoint.Call: error occurred: '%v'", err)
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
		return nil, vmtypes.ErrWrongTypeEntryPoint
	}
	return fun(ctx)
}

func (ep accountsEntryPoint) WithGasLimit(_ int) vmtypes.EntryPoint {
	return ep
}
