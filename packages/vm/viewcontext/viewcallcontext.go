package viewcontext

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
)

type callContext struct {
	contract iscp.Hname  // called contract
	params   iscp.Params // params passed
}

func (ctx *ViewContext) getCallContext() *callContext {
	if len(ctx.callStack) == 0 {
		panic("getCallContext: stack is empty")
	}
	return ctx.callStack[len(ctx.callStack)-1]
}

func (ctx *ViewContext) pushCallContext(contract iscp.Hname, params dict.Dict) {
	ctx.callStack = append(ctx.callStack,
		&callContext{
			contract: contract,
			params: iscp.Params{
				Dict:      params,
				KVDecoder: kvdecoder.New(params, ctx.log),
			},
		})
}

func (ctx *ViewContext) popCallContext() {
	ctx.callStack[len(ctx.callStack)-1] = nil // for GC
	ctx.callStack = ctx.callStack[:len(ctx.callStack)-1]
}
