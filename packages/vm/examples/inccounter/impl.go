package inccounter

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

type incCounterProcessor map[sctransaction.RequestCode]incEntryPoint

const (
	ProgramHash = "9qJQozz1TMhaJ2iYZUuxs49qL9LQYGJJ7xaVfE1TCf15"

	RequestInc                     = sctransaction.RequestCode(uint16(1))
	RequestIncAndRepeatOnceAfter5s = sctransaction.RequestCode(uint16(2))
	RequestIncAndRepeatMany        = sctransaction.RequestCode(uint16(3))

	ArgNumRepeats = "numrepeats"
	VarNumRepeats = "numrepeats"
	VarCounter    = "counter"
)

var entryPoints = incCounterProcessor{
	RequestInc:                     incCounter,
	RequestIncAndRepeatOnceAfter5s: incCounterAndRepeatOnce,
	RequestIncAndRepeatMany:        incCounterAndRepeatMany,
}

type incEntryPoint func(ctx vmtypes.Sandbox)

func GetProcessor() vmtypes.Processor {
	return entryPoints
}

func (proc incCounterProcessor) GetEntryPoint(rc sctransaction.RequestCode) (vmtypes.EntryPoint, bool) {
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

func (ep incEntryPoint) Run(ctx vmtypes.Sandbox) {
	ep(ctx)
}

func incCounter(ctx vmtypes.Sandbox) {
	state := ctx.AccessState()
	val, _ := state.GetInt64(VarCounter)
	ctx.Publish(fmt.Sprintf("'increasing counter value: %d'", val))
	state.SetInt64(VarCounter, val+1)
}

func incCounterAndRepeatOnce(ctx vmtypes.Sandbox) {
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
}

func incCounterAndRepeatMany(ctx vmtypes.Sandbox) {
	state := ctx.AccessState()

	val, _ := state.GetInt64(VarCounter)
	state.SetInt64(VarCounter, val+1)
	ctx.Publish(fmt.Sprintf("'increasing counter value: %d'", val))

	numRepeats, ok, err := ctx.AccessRequest().Args().GetInt64(ArgNumRepeats)
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
		return
	}

	ctx.Publishf("chain of %d requests ahead", numRepeats)

	state.SetInt64(VarNumRepeats, numRepeats-1)

	if ctx.SendRequestToSelfWithDelay(RequestIncAndRepeatMany, nil, 3) {
		ctx.Publishf("SendRequestToSelfWithDelay. remaining repeats = %d", numRepeats-1)
	} else {
		ctx.Publishf("SendRequestToSelfWithDelay FAILED. remaining repeats = %d", numRepeats-1)
	}
}
