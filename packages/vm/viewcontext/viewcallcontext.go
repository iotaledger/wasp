package viewcontext

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
)

type callContext struct {
	contract isc.Hname  // called contract
	params   isc.Params // params passed
}

func (ctx *ViewContext) getCallContext() *callContext {
	if len(ctx.callStack) == 0 {
		panic("getCallContext: stack is empty")
	}
	return ctx.callStack[len(ctx.callStack)-1]
}

func (ctx *ViewContext) pushCallContext(contract isc.Hname, params dict.Dict) {
	ctx.callStack = append(ctx.callStack,
		&callContext{
			contract: contract,
			params: isc.Params{
				Dict:      params,
				KVDecoder: kvdecoder.New(params, ctx.log),
			},
		})
}

func (ctx *ViewContext) popCallContext() {
	ctx.callStack[len(ctx.callStack)-1] = nil // for GC
	ctx.callStack = ctx.callStack[:len(ctx.callStack)-1]
}
