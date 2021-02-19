package inccounter

import (
	"fmt"
	"github.com/iotaledger/wasp/contracts/native"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

const (
	Name        = "inccounter"
	description = "Increment counter, a PoC smart contract"
)

var (
	Interface = &coreutil.ContractInterface{
		Name:        Name,
		Description: description,
		ProgramHash: hashing.HashStrings(Name),
	}
)

func init() {
	Interface.WithFunctions(initialize, []coreutil.ContractFunctionInterface{
		coreutil.Func(FuncIncCounter, incCounter),
		coreutil.Func(FuncIncAndRepeatOnceAfter5s, incCounterAndRepeatOnce),
		coreutil.Func(FuncIncAndRepeatMany, incCounterAndRepeatMany),
		coreutil.Func(FuncSpawn, spawn),
		coreutil.ViewFunc(FuncGetCounter, getCounter),
	})
}

const (
	FuncIncCounter              = "incCounter"
	FuncIncAndRepeatOnceAfter5s = "incAndRepeatOnceAfter5s"
	FuncIncAndRepeatMany        = "incAndRepeatMany"
	FuncSpawn                   = "spawn"
	FuncGetCounter              = "getCounter"
)

const (
	VarNumRepeats  = "numRepeats"
	VarCounter     = "counter"
	VarName        = "name"
	VarDescription = "dscr"
)

func init() {
	native.AddProcessor(Interface)
}

func initialize(ctx coretypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("inccounter.init in %s", ctx.ContractID().Hname().String())
	params := ctx.Params()
	val, _, err := codec.DecodeInt64(params.MustGet(VarCounter))
	if err != nil {
		return nil, fmt.Errorf("incCounter: %v", err)
	}
	ctx.State().Set(VarCounter, codec.EncodeInt64(val))
	ctx.Event(fmt.Sprintf("inccounter.init.success. counter = %d", val))
	return nil, nil
}

func incCounter(ctx coretypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("inccounter.incCounter in %s", ctx.ContractID().Hname().String())
	params := ctx.Params()
	inc, ok, err := codec.DecodeInt64(params.MustGet(VarCounter))
	if err != nil {
		return nil, err
	}
	if !ok {
		inc = 1
	}
	state := ctx.State()
	val, _, _ := codec.DecodeInt64(state.MustGet(VarCounter))
	ctx.Log().Debugf("incCounter: increasing counter value %d by %d", val, inc)
	state.Set(VarCounter, codec.EncodeInt64(val+inc))
	return nil, nil
}

func incCounterAndRepeatOnce(ctx coretypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("inccounter.incCounterAndRepeatOnce")
	state := ctx.State()
	val, _, _ := codec.DecodeInt64(state.MustGet(VarCounter))

	ctx.Log().Debugf(fmt.Sprintf("incCounterAndRepeatOnce: increasing counter value: %d", val))
	state.Set(VarCounter, codec.EncodeInt64(val+1))
	if !ctx.PostRequest(coretypes.PostRequestParams{
		TargetContractID: ctx.ContractID(),
		EntryPoint:       coretypes.Hn(FuncIncCounter),
		TimeLock:         5 * 60,
	}) {
		return nil, fmt.Errorf("incCounterAndRepeatOnce: not enough funds")
	}
	ctx.Log().Debugf("incCounterAndRepeatOnce: PostRequestToSelfWithDelay RequestInc 5 sec")
	return nil, nil
}

func incCounterAndRepeatMany(ctx coretypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("inccounter.incCounterAndRepeatMany")

	state := ctx.State()
	params := ctx.Params()

	val, _, _ := codec.DecodeInt64(state.MustGet(VarCounter))
	state.Set(VarCounter, codec.EncodeInt64(val+1))
	ctx.Log().Debugf("inccounter.incCounterAndRepeatMany: increasing counter value: %d", val)

	numRepeats, ok, err := codec.DecodeInt64(params.MustGet(VarNumRepeats))
	if err != nil {
		ctx.Log().Panicf("%s", err)
	}
	if !ok {
		numRepeats, _, _ = codec.DecodeInt64(state.MustGet(VarNumRepeats))
	}
	if numRepeats == 0 {
		ctx.Log().Debugf("inccounter.incCounterAndRepeatMany: finished chain of requests. counter value: %d", val)
		return nil, nil
	}

	ctx.Log().Debugf("chain of %d requests ahead", numRepeats)

	state.Set(VarNumRepeats, codec.EncodeInt64(numRepeats-1))

	if ctx.PostRequest(coretypes.PostRequestParams{
		TargetContractID: ctx.ContractID(),
		EntryPoint:       coretypes.Hn(FuncIncAndRepeatMany),
		TimeLock:         1 * 60,
	}) {
		ctx.Log().Debugf("PostRequestToSelfWithDelay. remaining repeats = %d", numRepeats-1)
	} else {
		ctx.Log().Debugf("PostRequestToSelfWithDelay FAILED. remaining repeats = %d", numRepeats-1)
	}
	return nil, nil
}

// spawn deploys new contract and calls it
func spawn(ctx coretypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("inccounter.spawn")
	state := ctx.State()

	val, _, _ := codec.DecodeInt64(state.MustGet(VarCounter))

	name, ok, err := codec.DecodeString(ctx.Params().MustGet(VarName))
	if err != nil {
		ctx.Log().Panicf("%v", err)
	}
	if !ok {
		return nil, fmt.Errorf("parameter 'name' wasnt found")
	}
	dscr, ok, err := codec.DecodeString(ctx.Params().MustGet(VarDescription))
	if err != nil {
		ctx.Log().Panicf("%v", err)
	}
	if !ok {
		dscr = "N/A"
	}
	par := dict.New()
	par.Set(VarCounter, codec.EncodeInt64(val+1))
	err = ctx.DeployContract(Interface.ProgramHash, name, dscr, par)
	if err != nil {
		return nil, err
	}

	// increase counter in newly spawned contract
	hname := coretypes.Hn(name)
	_, err = ctx.Call(hname, coretypes.Hn(FuncIncCounter), nil, nil)
	if err != nil {
		return nil, err
	}

	ctx.Log().Debugf("inccounter.spawn: new contract name = %s hname = %s", name, hname.String())
	return nil, nil
}

func getCounter(ctx coretypes.SandboxView) (dict.Dict, error) {
	state := ctx.State()
	val, _, _ := codec.DecodeInt64(state.MustGet(VarCounter))
	ret := dict.New()
	ret.Set(VarCounter, codec.EncodeInt64(val))
	return ret, nil
}
