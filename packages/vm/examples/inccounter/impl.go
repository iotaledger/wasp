package inccounter

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

type incCounterProcessor map[sctransaction.RequestCode]incEntryPoint

const (
	ProgramHash = "9qJQozz1TMhaJ2iYZUuxs49qL9LQYGJJ7xaVfE1TCf15"

	RequestInc          = sctransaction.RequestCode(uint16(1))
	RequestIncAndRepeat = sctransaction.RequestCode(uint16(2))
)

var entryPoints = incCounterProcessor{
	RequestInc:          incCounter,
	RequestIncAndRepeat: incCounterAndRepeat,
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

func incCounterAndRepeat(ctx vmtypes.Sandbox) {
	state := ctx.AccessState()
	val, _, _ := state.GetInt64("counter")

	ctx.Publish(fmt.Sprintf("'increasing counter value: %d'", val))
	state.SetInt64("counter", val+1)
	if val == 0 {
		ctx.GetWaspLog().Infof("[+++VM+++] val = %d", val)

		ctx.SendRequestToSelfWithDelay(RequestIncAndRepeat, nil, 3)
	}
}
