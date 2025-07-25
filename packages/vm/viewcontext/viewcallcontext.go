package viewcontext

import (
	"github.com/iotaledger/wasp/v2/packages/isc"
)

type callContext struct {
	contract isc.Hname         // called contract
	params   isc.CallArguments // params passed
}

func (ctx *ViewContext) getCallContext() *callContext {
	if len(ctx.callStack) == 0 {
		panic("getCallContext: stack is empty")
	}
	return ctx.callStack[len(ctx.callStack)-1]
}

func (ctx *ViewContext) pushCallContext(contract isc.Hname, params isc.CallArguments) {
	ctx.callStack = append(ctx.callStack,
		&callContext{
			contract: contract,
			params:   params,
		})
}

func (ctx *ViewContext) popCallContext() {
	ctx.callStack[len(ctx.callStack)-1] = nil // for GC
	ctx.callStack = ctx.callStack[:len(ctx.callStack)-1]
}
