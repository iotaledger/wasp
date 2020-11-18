package vmcontext

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

func (vmctx *VMContext) pushCallContextWithTransfer(contract coretypes.Hname, params codec.ImmutableCodec, transfer coretypes.ColoredBalances) error {
	if transfer != nil {
		agentID := coretypes.NewAgentIDFromContractID(coretypes.NewContractID(vmctx.ChainID(), contract))
		if len(vmctx.callStack) == 0 {
			vmctx.creditToAccount(agentID, transfer)
		} else {
			fromAgentID := coretypes.NewAgentIDFromContractID(coretypes.NewContractID(vmctx.ChainID(), vmctx.CurrentContractHname()))
			if !vmctx.moveBetweenAccounts(fromAgentID, agentID, transfer) {
				return fmt.Errorf("pushCallContextWithTransfer: transfer failed")
			}
		}
	}
	vmctx.pushCallContext(contract, params, transfer)
	return nil
}

func (vmctx *VMContext) pushCallContext(contract coretypes.Hname, params codec.ImmutableCodec, transfer coretypes.ColoredBalances) {
	vmctx.Log().Debugf("+++++++++++ PUSH %d, stack depth = %d", contract, len(vmctx.callStack))
	var caller coretypes.AgentID
	isRequestContext := len(vmctx.callStack) == 0
	if isRequestContext {
		// request context
		caller = vmctx.reqRef.SenderAgentID()
	} else {
		caller = coretypes.NewAgentIDFromContractID(vmctx.CurrentContractID())
	}
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
