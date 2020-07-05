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

func (ep incEntryPoint) WithGasLimit(gas int) vmtypes.EntryPoint {
	return ep
}

func (ep incEntryPoint) Run(ctx vmtypes.Sandbox) {
	ep(ctx)
}

func incCounter(ctx vmtypes.Sandbox) {
	state := ctx.AccessState()
	val, _, _ := state.GetInt64("counter")
	ctx.Publish(fmt.Sprintf("'increasing counter value: %d'", val))
	state.SetInt64("counter", val+1)
}

func incCounterAndRepeatOnce(ctx vmtypes.Sandbox) {
	state := ctx.AccessState()
	val, _, _ := state.GetInt64("counter")

	ctx.Publish(fmt.Sprintf("'increasing counter value: %d'", val))
	state.SetInt64("counter", val+1)
	if val == 0 {
		ctx.GetWaspLog().Info("SendRequestToSelfWithDelay 5 sec")

		ctx.SendRequestToSelfWithDelay(RequestIncAndRepeatOnceAfter5s, nil, 5)
	}
}

func incCounterAndRepeatMany(ctx vmtypes.Sandbox) {
	state := ctx.AccessState()

	val, _, _ := state.GetInt64("counter")
	state.SetInt64("counter", val+1)
	ctx.Publish(fmt.Sprintf("'increasing counter value: %d'", val))

	numRepeats, ok, err := ctx.AccessRequest().Args().GetInt64(ArgNumRepeats)
	if err != nil {
		ctx.Rollback()
		return
	}
	if !ok {
		numRepeats, ok, err = state.GetInt64(ArgNumRepeats)
		if err != nil {
			ctx.Rollback()
			return
		}
	}
	if numRepeats == 0 {
		ctx.GetWaspLog().Infof("finished chain of requests")
		return
	}

	ctx.GetWaspLog().Infof("chain of %d requests ahead", numRepeats)

	state.SetInt64(ArgNumRepeats, numRepeats-1)

	ctx.SendRequestToSelfWithDelay(RequestIncAndRepeatMany, nil, 3)
}
