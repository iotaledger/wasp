package inccounter

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
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
)

var entryPoints = incCounterProcessor{
	coretypes.EntryPointCodeInit:      initialize,
	EntryPointIncCounter:              incCounter,
	EntryPointIncAndRepeatOnceAfter5s: incCounterAndRepeatOnce,
	EntryPointIncAndRepeatMany:        incCounterAndRepeatMany,
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
		ctx.Publishf("error %v", err)
	}
	return nil, err
}

func initialize(ctx vmtypes.Sandbox) error {
	ctx.Publishf("inccounter.init")
	return nil
}

func incCounter(ctx vmtypes.Sandbox) error {
	state := ctx.AccessState()
	val, _ := state.GetInt64(VarCounter)
	ctx.Publish(fmt.Sprintf("'increasing counter value: %d'", val))
	state.SetInt64(VarCounter, val+1)
	return nil
}

func incCounterAndRepeatOnce(ctx vmtypes.Sandbox) error {
	state := ctx.AccessState()
	val, _ := state.GetInt64(VarCounter)

	ctx.Publish(fmt.Sprintf("increasing counter value: %d", val))
	state.SetInt64(VarCounter, val+1)
	if val == 0 {

		if ctx.SendRequestToSelfWithDelay(EntryPointIncCounter, nil, 5) {
			ctx.Publish("SendRequestToSelfWithDelay RequestInc 5 sec")
		} else {
			ctx.Publish("failed to SendRequestToSelfWithDelay RequestInc 5 sec")
		}
	}
	return nil
}

func incCounterAndRepeatMany(ctx vmtypes.Sandbox) error {
	state := ctx.AccessState()
	params := ctx.Params()

	val, _ := state.GetInt64(VarCounter)
	state.SetInt64(VarCounter, val+1)
	ctx.Publish(fmt.Sprintf("'increasing counter value: %d'", val))

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
		ctx.GetWaspLog().Infof("finished chain of requests")
		return nil
	}

	ctx.Publishf("chain of %d requests ahead", numRepeats)

	state.SetInt64(VarNumRepeats, numRepeats-1)

	if ctx.SendRequestToSelfWithDelay(EntryPointIncAndRepeatMany, nil, 3) {
		ctx.Publishf("SendRequestToSelfWithDelay. remaining repeats = %d", numRepeats-1)
	} else {
		ctx.Publishf("SendRequestToSelfWithDelay FAILED. remaining repeats = %d", numRepeats-1)
	}
	return nil
}
