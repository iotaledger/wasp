package inccounter

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

type incCounterProcessor map[coretypes.Hname]incEntryPoint

const (
	ProgramHashStr = "9qJQozz1TMhaJ2iYZUuxs49qL9LQYGJJ7xaVfE1TCf15"
	Description    = "Increment counter, a PoC smart contract"

	VarNumRepeats  = "numRepeats"
	VarCounter     = "counter"
	VarName        = "name"
	VarDescription = "dscr"
)

var (
	EntryPointIncCounter              = coretypes.Hn("incCounter")
	EntryPointIncAndRepeatOnceAfter5s = coretypes.Hn("incAndRepeatOnceAfter5s")
	EntryPointIncAndRepeatMany        = coretypes.Hn("incAndRepeatMany")
	EntryPointSpawn                   = coretypes.Hn("spawn")
	EntryPointGetCounter              = coretypes.Hn("getCounter")

	ProgramHash, _ = hashing.HashValueFromBase58(ProgramHashStr)
)

var entryPoints = incCounterProcessor{
	coretypes.EntryPointInit:          epFunc(initialize),
	EntryPointIncCounter:              epFunc(incCounter),
	EntryPointIncAndRepeatOnceAfter5s: epFunc(incCounterAndRepeatOnce),
	EntryPointIncAndRepeatMany:        epFunc(incCounterAndRepeatMany),
	EntryPointSpawn:                   epFunc(spawn),
	EntryPointGetCounter:              epView(getCounter),
}

type incEntryPoint struct {
	view    bool
	funFull func(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error)
	funView func(ctx vmtypes.SandboxView) (codec.ImmutableCodec, error)
}

func epFunc(fun func(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error)) incEntryPoint {
	return incEntryPoint{funFull: fun}
}

func epView(fun func(ctx vmtypes.SandboxView) (codec.ImmutableCodec, error)) incEntryPoint {
	return incEntryPoint{funView: fun, view: true}
}

func GetProcessor() vmtypes.Processor {
	return entryPoints
}

func (proc incCounterProcessor) GetEntryPoint(rc coretypes.Hname) (vmtypes.EntryPoint, bool) {
	f, ok := proc[rc]
	if !ok {
		return nil, false
	}
	return f, true
}

func (v incCounterProcessor) GetDescription() string {
	return "IncrementCounter hard coded smart contract processor"
}

func (ep incEntryPoint) WithGasLimit(gas int) vmtypes.EntryPoint {
	return ep
}

func (ep incEntryPoint) IsView() bool {
	return ep.view
}

func (ep incEntryPoint) Call(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	if ep.IsView() {
		panic("wrong call of the view")
	}
	return ep.funFull(ctx)
}

func (ep incEntryPoint) CallView(ctx vmtypes.SandboxView) (codec.ImmutableCodec, error) {
	if !ep.IsView() {
		panic("wrong call of the full entry point")
	}
	return ep.funView(ctx)
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

		if ctx.PostRequestToSelfWithDelay(EntryPointIncCounter, nil, 5) {
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
		numRepeats, ok = state.GetInt64(VarNumRepeats)
		if err != nil {
			ctx.Panic(err)
		}
	}
	if numRepeats == 0 {
		ctx.Eventf("inccounter.incCounterAndRepeatMany: finished chain of requests. counter value: %d", val)
		return nil, nil
	}

	ctx.Eventf("chain of %d requests ahead", numRepeats)

	state.SetInt64(VarNumRepeats, numRepeats-1)

	if ctx.PostRequestToSelfWithDelay(EntryPointIncAndRepeatMany, nil, 1) {
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
	hname := coretypes.Hn(name)
	_, err = ctx.Call(hname, EntryPointIncCounter, nil, nil)
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
