package increasecounter

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

type increaseCounterProcessor map[sctransaction.RequestCode]increaseEntryPoint

const (
	ProgramHash = "9qJQozz1TMhaJ2iYZUuxs49qL9LQYGJJ7xaVfE1TCf15"

	RequestIncrease = sctransaction.RequestCode(uint16(1))
)

var entryPoints = increaseCounterProcessor{
	RequestIncrease: increaseCounter,
}

type increaseEntryPoint func(ctx vmtypes.Sandbox)

func GetProcessor() vmtypes.Processor {
	return entryPoints
}

func (f increaseCounterProcessor) GetEntryPoint(_ sctransaction.RequestCode) (vmtypes.EntryPoint, bool) {
	return (increaseEntryPoint)(increaseCounter), true
}

func (ep increaseEntryPoint) WithGasLimit(gas int) vmtypes.EntryPoint {
	return ep
}

func (ep increaseEntryPoint) Run(ctx vmtypes.Sandbox) {
	ep(ctx)
}

func increaseCounter(ctx vmtypes.Sandbox) {
	val, _, _ := ctx.AccessState().Variables().GetInt64("counter")
	ctx.Publish(fmt.Sprintf("'increasing counter value: %d'", val))
	ctx.AccessState().Variables().SetInt64("counter", val+1)
}
