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
	ProgramHash = "9qJQozz1TMhaJ2iYZUuxs49qL9LQYGJJ7xaVfE1TCf15"
	Description = "Increment counter, a PoC smart contract"

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
)

var entryPoints = incCounterProcessor{
	coretypes.EntryPointCodeInit:      initialize,
	EntryPointIncCounter:              incCounter,
	EntryPointIncAndRepeatOnceAfter5s: incCounterAndRepeatOnce,
	EntryPointIncAndRepeatMany:        incCounterAndRepeatMany,
	EntryPointSpawn:                   spawn,
}

type incEntryPoint func(ctx vmtypes.Sandbox) error

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

func (ep incEntryPoint) Call(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	err := ep(ctx)
	if err != nil {
		ctx.Eventf("error %v", err)
	}
	return nil, err
}

// TODO
func (ep incEntryPoint) IsView() bool {
	return false
}

// TODO
func (ep incEntryPoint) CallView(ctx vmtypes.SandboxView) (codec.ImmutableCodec, error) {
	panic("implement me")
}

func initialize(ctx vmtypes.Sandbox) error {
	ctx.Eventf("inccounter.init")
	params := ctx.Params()
	val, _, err := params.GetInt64(VarCounter)
	if err != nil {
		return fmt.Errorf("incCounter: %v", err)
	}
	ctx.AccessState().SetInt64(VarCounter, val)
	ctx.Eventf("inccounter.init.success. counter = %d", val)
	return nil
}

func incCounter(ctx vmtypes.Sandbox) error {
	ctx.Eventf("inccounter.incCounter")
	state := ctx.AccessState()
	val, _ := state.GetInt64(VarCounter)
	ctx.Event(fmt.Sprintf("'increasing counter value: %d'", val))
	state.SetInt64(VarCounter, val+1)
	return nil
}

func incCounterAndRepeatOnce(ctx vmtypes.Sandbox) error {
	ctx.Eventf("inccounter.incCounterAndRepeatOnce")
	state := ctx.AccessState()
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
	return nil
}

func incCounterAndRepeatMany(ctx vmtypes.Sandbox) error {
	ctx.Eventf("inccounter.incCounterAndRepeatMany")

	state := ctx.AccessState()
	params := ctx.Params()

	val, _ := state.GetInt64(VarCounter)
	state.SetInt64(VarCounter, val+1)
	ctx.Event(fmt.Sprintf("'increasing counter value: %d'", val))

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
		ctx.Eventf("finished chain of requests")
		return nil
	}

	ctx.Eventf("chain of %d requests ahead", numRepeats)

	state.SetInt64(VarNumRepeats, numRepeats-1)

	if ctx.PostRequestToSelfWithDelay(EntryPointIncAndRepeatMany, nil, 3) {
		ctx.Eventf("PostRequestToSelfWithDelay. remaining repeats = %d", numRepeats-1)
	} else {
		ctx.Eventf("PostRequestToSelfWithDelay FAILED. remaining repeats = %d", numRepeats-1)
	}
	return nil
}

// spawn deploys new contract an calls it
func spawn(ctx vmtypes.Sandbox) error {
	ctx.Eventf("inccounter.spawn")
	state := ctx.AccessState()

	val, _ := state.GetInt64(VarCounter)

	hashBin, err := hashing.HashValueFromBase58(ProgramHash)
	if err != nil {
		ctx.Panic(err)
	}
	name, ok, err := ctx.Params().GetString(VarName)
	if err != nil {
		ctx.Panic(err)
	}
	if !ok {
		return fmt.Errorf("parameter 'name' wasnt found")
	}
	dscr, ok, err := ctx.Params().GetString(VarDescription)
	if err != nil {
		ctx.Panic(err)
	}
	if !ok {
		dscr = "N/A"
	}
	par := codec.NewCodec(dict.NewDict())
	par.SetInt64(VarCounter, val+1)
	err = ctx.DeployContract("examplevm", hashBin[:], name, dscr, par)
	if err != nil {
		return err
	}

	// increase counter in newly spawned contract
	hname := coretypes.Hn(name)
	_, err = ctx.Call(hname, EntryPointIncCounter, nil, nil)
	if err != nil {
		return err
	}

	ctx.Eventf("inccounter.spawn: new contract name = %s hname = %s", name, hname.String())
	return nil
}
