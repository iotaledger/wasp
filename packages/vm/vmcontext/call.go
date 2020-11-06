package vmcontext

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

// CallContract
func (vmctx *VMContext) CallContract(contractIndex uint16, epCode coretypes.EntryPointCode, params codec.ImmutableCodec, budget map[balance.Color]int64) (codec.ImmutableCodec, error) {
	vmctx.log.Debugw("CallContract", "contactIndex", contractIndex, "epCode", epCode.String())

	rec, ok := vmctx.findContractByIndex(contractIndex)
	if !ok {
		return nil, fmt.Errorf("failed to find contract with index %d", contractIndex)
	}

	proc, err := vmctx.getProcessor(rec)
	if err != nil {
		return nil, err
	}

	ep, ok := proc.GetEntryPoint(epCode)
	if !ok {
		return nil, fmt.Errorf("can't find entry point for entry point '%s'", epCode.String())
	}

	if err := vmctx.PushCallContext(contractIndex, params, budget); err != nil {
		return nil, err
	}
	defer vmctx.PopCallContext()

	return ep.Call(NewSandbox(vmctx))
}

func (vmctx *VMContext) callFromRequest() {
	req := vmctx.reqRef.RequestSection()
	vmctx.log.Debugf("callFromRequest: %s -- %s\n", vmctx.reqRef.RequestID().String(), req.String())

	_, err := vmctx.CallContract(req.Target().Index(), req.EntryPointCode(), req.Args(), nil)
	if err != nil {
		vmctx.log.Warnf("callFromRequest: %v", err)
	}
}

func (vmctx *VMContext) Params() codec.ImmutableCodec {
	return vmctx.callStack[len(vmctx.callStack)-1].params
}
