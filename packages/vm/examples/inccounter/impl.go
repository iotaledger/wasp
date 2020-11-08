package inccounter

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

type incCounterProcessor map[coretypes.EntryPointCode]incEntryPoint

const (
	ProgramHash = "9qJQozz1TMhaJ2iYZUuxs49qL9LQYGJJ7xaVfE1TCf15"
	Description = "Increment counter, a PoC smart contract"

	VarNumRepeats = "numRepeats"
	VarCounter    = "counter"
)

var (
	EntryPointIncCounter              = coretypes.NewEntryPointCodeFromFunctionName("incCounter")
	EntryPointIncAndRepeatOnceAfter5s = coretypes.NewEntryPointCodeFromFunctionName("incAndRepeatOnceAfter5s")
	EntryPointIncAndRepeatMany        = coretypes.NewEntryPointCodeFromFunctionName("incAndRepeatMany")
	EntryPointSpawn                   = coretypes.NewEntryPointCodeFromFunctionName("spawn")
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

func (proc incCounterProcessor) GetEntryPoint(rc coretypes.EntryPointCode) (vmtypes.EntryPoint, bool) {
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

		if ctx.SendRequestToSelfWithDelay(EntryPointIncCounter, nil, 5) {
			ctx.Event("SendRequestToSelfWithDelay RequestInc 5 sec")
		} else {
			ctx.Event("failed to SendRequestToSelfWithDelay RequestInc 5 sec")
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

	if ctx.SendRequestToSelfWithDelay(EntryPointIncAndRepeatMany, nil, 3) {
		ctx.Eventf("SendRequestToSelfWithDelay. remaining repeats = %d", numRepeats-1)
	} else {
		ctx.Eventf("SendRequestToSelfWithDelay FAILED. remaining repeats = %d", numRepeats-1)
	}
	return nil
}

func spawn(ctx vmtypes.Sandbox) error {
	ctx.Eventf("inccounter.spawn")
	state := ctx.AccessState()

	val, _ := state.GetInt64(VarCounter)

	hashBin, err := hashing.HashValueFromBase58(ProgramHash)
	if err != nil {
		ctx.Panic(err)
	}
	par := codec.NewCodec(dict.NewDict())
	par.SetInt64(VarCounter, val+1)
	spawnedContractIndex, err := ctx.DeployContract("examplevm", hashBin[:], "", "Inccounter spawned", par)
	if err != nil {
		return err
	}
	ctx.Eventf("inccounter.spawn: new contract index = %d", spawnedContractIndex)
	return nil
}
