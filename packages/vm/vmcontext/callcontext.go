package vmcontext

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coret"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

func (vmctx *VMContext) pushCallContextWithTransfer(contract coret.Hname, params codec.ImmutableCodec, transfer coret.ColoredBalances) error {
	if transfer != nil {
		agentID := coret.NewAgentIDFromContractID(coret.NewContractID(vmctx.ChainID(), contract))
		if len(vmctx.callStack) == 0 {
			vmctx.creditToAccount(agentID, transfer)
		} else {
			fromAgentID := coret.NewAgentIDFromContractID(coret.NewContractID(vmctx.ChainID(), vmctx.CurrentContractHname()))
			if !vmctx.moveBetweenAccounts(fromAgentID, agentID, transfer) {
				return fmt.Errorf("pushCallContextWithTransfer: transfer failed")
			}
		}
	}
	vmctx.pushCallContext(contract, params, transfer)
	return nil
}

func (vmctx *VMContext) pushCallContext(contract coret.Hname, params codec.ImmutableCodec, transfer coret.ColoredBalances) {
	vmctx.Log().Debugf("+++++++++++ PUSH %d, stack depth = %d", contract, len(vmctx.callStack))
	var caller coret.AgentID
	isRequestContext := len(vmctx.callStack) == 0
	if isRequestContext {
		// request context
		caller = vmctx.reqRef.SenderAgentID()
	} else {
		caller = coret.NewAgentIDFromContractID(vmctx.CurrentContractID())
	}
	vmctx.Log().Debugf("+++++++++++ PUSH %d, stack depth = %d caller = %s", contract, len(vmctx.callStack), caller.String())
	vmctx.callStack = append(vmctx.callStack, &callContext{
		isRequestContext: isRequestContext,
		caller:           caller,
		contract:         contract,
		params:           params,
		transfer:         transfer,
	})
}

func (vmctx *VMContext) popCallContext() {
	vmctx.Log().Debugf("+++++++++++ POP @ depth %d", len(vmctx.callStack))
	vmctx.callStack = vmctx.callStack[:len(vmctx.callStack)-1]
}

func (vmctx *VMContext) getCallContext() *callContext {
	if len(vmctx.callStack) == 0 {
		panic("getCallContext: stack is empty")
	}
	return vmctx.callStack[len(vmctx.callStack)-1]
}
