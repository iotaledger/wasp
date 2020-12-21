package inccounter

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/contract"
	"github.com/iotaledger/wasp/packages/vm/examples"
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

func init() {
	examples.AddProcessor(ProgramHash, Interface)
}

func GetProcessor() vmtypes.Processor {
	return Interface
}

func initialize(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Eventf("inccounter.init in %s", ctx.ContractID().Hname().String())
	params := ctx.Params()
	val, _, err := codec.DecodeInt64(params.MustGet(VarCounter))
	if err != nil {
		return nil, fmt.Errorf("incCounter: %v", err)
	}
	ctx.State().Set(VarCounter, codec.EncodeInt64(val))
	ctx.Eventf("inccounter.init.success. counter = %d", val)
	return nil, nil
}

func incCounter(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Eventf("inccounter.incCounter in %s", ctx.ContractID().Hname().String())
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
	ctx.Eventf("incCounter: increasing counter value %d by %d", val, inc)
	state.Set(VarCounter, codec.EncodeInt64(val+inc))
	return nil, nil
}

func incCounterAndRepeatOnce(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Eventf("inccounter.incCounterAndRepeatOnce")
	state := ctx.State()
	val, _, _ := codec.DecodeInt64(state.MustGet(VarCounter))

	ctx.Event(fmt.Sprintf("incCounterAndRepeatOnce: increasing counter value: %d", val))
	state.Set(VarCounter, codec.EncodeInt64(val+1))
	if !ctx.PostRequest(vmtypes.NewRequestParams{
		TargetContractID: ctx.ContractID(),
		EntryPoint:       coretypes.Hn(FuncIncCounter),
		Timelock:         5 * 60,
	}) {
		return nil, fmt.Errorf("incCounterAndRepeatOnce: not enough funds")
	}
	ctx.Event("incCounterAndRepeatOnce: PostRequestToSelfWithDelay RequestInc 5 sec")
	return nil, nil
}

func incCounterAndRepeatMany(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Eventf("inccounter.incCounterAndRepeatMany")

	state := ctx.State()
	params := ctx.Params()

	val, _, _ := codec.DecodeInt64(state.MustGet(VarCounter))
	state.Set(VarCounter, codec.EncodeInt64(val+1))
	ctx.Eventf("inccounter.incCounterAndRepeatMany: increasing counter value: %d", val)

	numRepeats, ok, err := codec.DecodeInt64(params.MustGet(VarNumRepeats))
	if err != nil {
		ctx.Log().Panicf("%s", err)
	}
	if !ok {
		numRepeats, _, _ = codec.DecodeInt64(state.MustGet(VarNumRepeats))
	}
	if numRepeats == 0 {
		ctx.Eventf("inccounter.incCounterAndRepeatMany: finished chain of requests. counter value: %d", val)
		return nil, nil
	}

	ctx.Eventf("chain of %d requests ahead", numRepeats)

	state.Set(VarNumRepeats, codec.EncodeInt64(numRepeats-1))

	if ctx.PostRequest(vmtypes.NewRequestParams{
		TargetContractID: ctx.ContractID(),
		EntryPoint:       coretypes.Hn(FuncIncAndRepeatMany),
		Timelock:         1 * 60,
	}) {
		ctx.Eventf("PostRequestToSelfWithDelay. remaining repeats = %d", numRepeats-1)
	} else {
		ctx.Eventf("PostRequestToSelfWithDelay FAILED. remaining repeats = %d", numRepeats-1)
	}
	return nil, nil
}

// spawn deploys new contract and calls it
func spawn(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Eventf("inccounter.spawn")
	state := ctx.State()

	val, _, _ := codec.DecodeInt64(state.MustGet(VarCounter))

	hashBin, err := hashing.HashValueFromBase58(ProgramHashStr)
	if err != nil {
		ctx.Log().Panicf("%v", err)
	}
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
	err = ctx.DeployContract(hashBin, name, dscr, par)
	if err != nil {
		return nil, err
	}

	// increase counter in newly spawned contract
	hname := coretypes.Hn(name)
	_, err = ctx.Call(hname, coretypes.Hn(FuncIncCounter), nil, nil)
	if err != nil {
		return nil, err
	}

	ctx.Eventf("inccounter.spawn: new contract name = %s hname = %s", name, hname.String())
	return nil, nil
}

func getCounter(ctx vmtypes.SandboxView) (dict.Dict, error) {
	state := ctx.State()
	val, _, _ := codec.DecodeInt64(state.MustGet(VarCounter))
	ret := dict.New()
	ret.Set(VarCounter, codec.EncodeInt64(val))
	return ret, nil
}
