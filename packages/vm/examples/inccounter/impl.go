package inccounter

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

type incCounterProcessor map[coretypes.EntryPointCode]incEntryPoint

const (
	ProgramHash = "9qJQozz1TMhaJ2iYZUuxs49qL9LQYGJJ7xaVfE1TCf15"
	Description = "Increment, a PoC smart contract"

	RequestInc                     = coretypes.EntryPointCode(1)
	RequestIncAndRepeatOnceAfter5s = coretypes.EntryPointCode(2)
	RequestIncAndRepeatMany        = coretypes.EntryPointCode(3)

	ArgNumRepeats = "numRepeats"
	VarNumRepeats = "numRepeats"
	VarCounter    = "counter"
)

var entryPoints = incCounterProcessor{
	RequestInc:                     incCounter,
	RequestIncAndRepeatOnceAfter5s: incCounterAndRepeatOnce,
	RequestIncAndRepeatMany:        incCounterAndRepeatMany,
}

type incEntryPoint func(ctx vmtypes.Sandbox, params kv.RCodec) error

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

func (ep incEntryPoint) Call(ctx vmtypes.Sandbox, params kv.RCodec) interface{} {
	err := ep(ctx, params)
	if err != nil {
		ctx.Publishf("error %v", err)
	}
	return err
}

func incCounter(ctx vmtypes.Sandbox, _ kv.RCodec) error {
	state := ctx.AccessState()
	val, _ := state.GetInt64(VarCounter)
	ctx.Publish(fmt.Sprintf("'increasing counter value: %d'", val))
	state.SetInt64(VarCounter, val+1)
	return nil
}

func incCounterAndRepeatOnce(ctx vmtypes.Sandbox, _ kv.RCodec) error {
	state := ctx.AccessState()
	val, _ := state.GetInt64(VarCounter)

	ctx.Publish(fmt.Sprintf("increasing counter value: %d", val))
	state.SetInt64(VarCounter, val+1)
	if val == 0 {

		if ctx.SendRequestToSelfWithDelay(RequestInc, nil, 5) {
			ctx.Publish("SendRequestToSelfWithDelay RequestInc 5 sec")
		} else {
			ctx.Publish("failed to SendRequestToSelfWithDelay RequestInc 5 sec")
		}
	}
	return nil
}

func incCounterAndRepeatMany(ctx vmtypes.Sandbox, params kv.RCodec) error {
	state := ctx.AccessState()

	val, _ := state.GetInt64(VarCounter)
	state.SetInt64(VarCounter, val+1)
	ctx.Publish(fmt.Sprintf("'increasing counter value: %d'", val))

	numRepeats, ok, err := params.GetInt64(ArgNumRepeats)
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

	if ctx.SendRequestToSelfWithDelay(RequestIncAndRepeatMany, nil, 3) {
		ctx.Publishf("SendRequestToSelfWithDelay. remaining repeats = %d", numRepeats-1)
	} else {
		ctx.Publishf("SendRequestToSelfWithDelay FAILED. remaining repeats = %d", numRepeats-1)
	}
	return nil
}
