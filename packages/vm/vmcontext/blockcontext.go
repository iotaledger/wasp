package vmcontext

import "github.com/iotaledger/wasp/packages/iscp"

func (vmctx *VMContext) BlockContext(ctx iscp.Sandbox, construct func(ctx iscp.Sandbox) interface{}, onClose func(interface{})) interface{} {
	hname := vmctx.CurrentContractHname()
	if bctx, alreadyExists := vmctx.blockContext[hname]; alreadyExists {
		return bctx
	}
	if onClose == nil {
		onClose = func(interface{}) {}
	}
	ret := &blockContext{
		obj:     construct(ctx),
		onClose: onClose,
	}
	vmctx.blockContext[hname] = ret
	// storing sequence to have deterministic order of closing
	vmctx.blockContextCloseSeq = append(vmctx.blockContextCloseSeq, hname)
	return ret.obj
}
