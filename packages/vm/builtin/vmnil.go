package builtin

import (
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm"
)

type builtinProcessor struct {
}

type builtinEntryPoint struct {
}

var Processor = New()

func New() vm.Processor {
	return &builtinProcessor{}
}

func (v *builtinProcessor) GetEntryPoint(code sctransaction.RequestCode) (vm.EntryPoint, bool) {
	if !code.IsReserved() {
		return nil, false
	}
	return &builtinEntryPoint{}, true
}

func (v *builtinEntryPoint) Run(ctx vm.Sandbox) {
	reqId := ctx.GetRequestID()
	ctx.GetLog().Debugw("run nilProcessor",
		"request code", ctx.GetRequestCode(),
		"addr", ctx.GetAddress().String(),
		"ts", ctx.GetTimestamp(),
		"state index", ctx.GetStateIndex(),
		"req", reqId.String(),
	)
}
