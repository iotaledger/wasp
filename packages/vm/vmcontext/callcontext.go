package vmcontext

import (
	"fmt"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/sctransaction"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

func (vmctx *VMContext) pushCallContextWithTransfer(contract coretypes.Hname, params dict.Dict, transfer *ledgerstate.ColoredBalances) error {
	if transfer != nil {
		agentID := coretypes.NewAgentID(vmctx.ChainID().AsAddress(), contract)
		agentID = vmctx.adjustAccount(agentID)
		if len(vmctx.callStack) == 0 {
			// was this an off-ledger request?
			if _, ok := vmctx.req.(*sctransaction.RequestOffLedger); ok {
				sender := vmctx.req.SenderAccount()
				if !vmctx.moveBetweenAccounts(sender, agentID, transfer) {
					return fmt.Errorf("pushCallContextWithTransfer: off-ledger transfer failed: not enough funds")
				}
			} else {
				vmctx.creditToAccount(agentID, transfer)
			}
		} else {
			fromAgentID := coretypes.NewAgentID(vmctx.ChainID().AsAddress(), vmctx.CurrentContractHname())
			fromAgentID = vmctx.adjustAccount(fromAgentID)
			if !vmctx.moveBetweenAccounts(fromAgentID, agentID, transfer) {
				return fmt.Errorf("pushCallContextWithTransfer: transfer failed: not enough funds")
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
		caller = vmctx.req.SenderAccount()
	} else {
		caller = vmctx.MyAgentID()
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
