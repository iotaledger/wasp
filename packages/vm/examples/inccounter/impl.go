package inccounter

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/coret"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/contract"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

const (
	Name        = "inccounter"
	Version     = "0.1"
	fullName    = Name + "-" + Version
	description = "Increment counter, a PoC smart contract"
)

var (
	Interface = &contract.ContractInterface{
		Name:        fullName,
		Description: description,
		ProgramHash: *hashing.HashStrings(fullName),
	}
)

func init() {
	Interface.WithFunctions(initialize, []contract.ContractFunctionInterface{
		contract.Func(FuncIncCounter, incCounter),
		contract.Func(FuncIncAndRepeatOnceAfter5s, incCounterAndRepeatOnce),
		contract.Func(FuncIncAndRepeatMany, incCounterAndRepeatMany),
		contract.Func(FuncSpawn, spawn),
		contract.ViewFunc(FuncGetCounter, getCounter),
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
	ProgramHashStr = "9qJQozz1TMhaJ2iYZUuxs49qL9LQYGJJ7xaVfE1TCf15"

	VarNumRepeats  = "numRepeats"
	VarCounter     = "counter"
	VarName        = "name"
	VarDescription = "dscr"
)

var (
	ProgramHash, _ = hashing.HashValueFromBase58(ProgramHashStr)
)

func GetProcessor() vmtypes.Processor {
	return Interface
}

func initialize(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	ctx.Eventf("inccounter.init in %s", ctx.MyContractID().Hname().String())
	params := ctx.Params()
	val, _, err := params.GetInt64(VarCounter)
	if err != nil {
		return nil, fmt.Errorf("incCounter: %v", err)
	}
	ctx.State().SetInt64(VarCounter, val)
	ctx.Eventf("inccounter.init.success. counter = %d", val)
	return nil, nil
}

func incCounter(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	ctx.Eventf("inccounter.incCounter in %s", ctx.MyContractID().Hname().String())
	state := ctx.State()
	val, _ := state.GetInt64(VarCounter)
	ctx.Eventf("'increasing counter value: %d' in %s", val, ctx.MyContractID().Hname().String())
	state.SetInt64(VarCounter, val+1)
	return nil, nil
}

func incCounterAndRepeatOnce(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	ctx.Eventf("inccounter.incCounterAndRepeatOnce")
	state := ctx.State()
	val, _ := state.GetInt64(VarCounter)

	ctx.Event(fmt.Sprintf("increasing counter value: %d", val))
	state.SetInt64(VarCounter, val+1)
	if val == 0 {
		if ctx.PostRequestToSelfWithDelay(coret.Hn(FuncIncCounter), nil, 5) {
			ctx.Event("PostRequestToSelfWithDelay RequestInc 5 sec")
		} else {
			ctx.Event("failed to PostRequestToSelfWithDelay RequestInc 5 sec")
		}
	}
	return nil, nil
}

func incCounterAndRepeatMany(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	ctx.Eventf("inccounter.incCounterAndRepeatMany")

	state := ctx.State()
	params := ctx.Params()

	val, _ := state.GetInt64(VarCounter)
	state.SetInt64(VarCounter, val+1)
	ctx.Eventf("inccounter.incCounterAndRepeatMany: increasing counter value: %d", val)

	numRepeats, ok, err := params.GetInt64(VarNumRepeats)
	if err != nil {
		ctx.Panic(err)
	}
	if !ok {
		numRepeats, _ = state.GetInt64(VarNumRepeats)
	}
	if numRepeats == 0 {
		ctx.Eventf("inccounter.incCounterAndRepeatMany: finished chain of requests. counter value: %d", val)
		return nil, nil
	}

	ctx.Eventf("chain of %d requests ahead", numRepeats)

	state.SetInt64(VarNumRepeats, numRepeats-1)

	if ctx.PostRequestToSelfWithDelay(coret.Hn(FuncIncAndRepeatMany), nil, 1) {
		ctx.Eventf("PostRequestToSelfWithDelay. remaining repeats = %d", numRepeats-1)
	} else {
		ctx.Eventf("PostRequestToSelfWithDelay FAILED. remaining repeats = %d", numRepeats-1)
	}
	return nil, nil
}

// spawn deploys new contract and calls it
func spawn(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	ctx.Eventf("inccounter.spawn")
	state := ctx.State()

	val, _ := state.GetInt64(VarCounter)

	hashBin, err := hashing.HashValueFromBase58(ProgramHashStr)
	if err != nil {
		ctx.Panic(err)
	}
	name, ok, err := ctx.Params().GetString(VarName)
	if err != nil {
		ctx.Panic(err)
	}
	if !ok {
		return nil, fmt.Errorf("parameter 'name' wasnt found")
	}
	dscr, ok, err := ctx.Params().GetString(VarDescription)
	if err != nil {
		ctx.Panic(err)
	}
	if !ok {
		dscr = "N/A"
	}
	par := codec.NewCodec(dict.New())
	par.SetInt64(VarCounter, val+1)
	err = ctx.CreateContract(hashBin, name, dscr, par)
	if err != nil {
		return nil, err
	}

	// increase counter in newly spawned contract
	hname := coret.Hn(name)
	_, err = ctx.Call(hname, coret.Hn(FuncIncCounter), nil, nil)
	if err != nil {
		return nil, err
	}

	ctx.Eventf("inccounter.spawn: new contract name = %s hname = %s", name, hname.String())
	return nil, nil
}

func getCounter(ctx vmtypes.SandboxView) (codec.ImmutableCodec, error) {
	val, _ := ctx.State().GetInt64(VarCounter)
	ret := codec.NewCodec(dict.New())
	ret.SetInt64(VarCounter, val)
	return ret, nil
}
