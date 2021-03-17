package vmcontext

import (
	"fmt"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

func (vmctx *VMContext) pushCallContextWithTransfer(contract coretypes.Hname, params dict.Dict, transfer *ledgerstate.ColoredBalances) error {
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

const traceStack = false

func (vmctx *VMContext) pushCallContext(contract coretypes.Hname, params dict.Dict, transfer *ledgerstate.ColoredBalances) {
	if traceStack {
		vmctx.log.Debugf("+++++++++++ PUSH %d, stack depth = %d", contract, len(vmctx.callStack))
	}
	var caller *coretypes.AgentID
	isRequestContext := len(vmctx.callStack) == 0
	if isRequestContext {
		// request context
		caller = vmctx.req.SenderAgentID()
	} else {
		caller = coretypes.NewAgentIDFromContractID(vmctx.CurrentContractID())
	}
	if traceStack {
		vmctx.log.Debugf("+++++++++++ PUSH %d, stack depth = %d caller = %s", contract, len(vmctx.callStack), caller.String())
	}
	if transfer != nil {
		transfer = transfer.Clone()
	}
	vmctx.callStack = append(vmctx.callStack, &callContext{
		isRequestContext: isRequestContext,
		caller:           *caller.Clone(),
		contract:         contract,
		params:           params.Clone(),
		transfer:         transfer,
	})
}

func (vmctx *VMContext) popCallContext() {
	if traceStack {
		vmctx.log.Debugf("+++++++++++ POP @ depth %d", len(vmctx.callStack))
	}
	vmctx.callStack[len(vmctx.callStack)-1] = nil // for GC
	vmctx.callStack = vmctx.callStack[:len(vmctx.callStack)-1]
}

func (vmctx *VMContext) getCallContext() *callContext {
	if len(vmctx.callStack) == 0 {
		panic("getCallContext: stack is empty")
	}
	return vmctx.callStack[len(vmctx.callStack)-1]
}
