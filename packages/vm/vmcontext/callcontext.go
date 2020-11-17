package vmcontext

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
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

// creditToAccount deposits transfer from request to chain account of of the called contract
// It adds new tokens to the chain ledger
// It is used when new tokens arrive with a request
func (vmctx *VMContext) creditToAccount(agentID coretypes.AgentID, transfer coretypes.ColoredBalances) {
	if len(vmctx.callStack) > 0 {
		vmctx.log.Panicf("creditToAccount must be called only from request")
	}
	vmctx.pushCallContext(accountsc.Hname, nil, nil) // create local context for the state
	defer vmctx.popCallContext()

	accountsc.CreditToAccount(codec.NewMustCodec(vmctx), agentID, transfer)
}

// debitFromAccount subtracts tokens from account if it is enough of it.
// should be called only when posting request
func (vmctx *VMContext) debitFromAccount(agentID coretypes.AgentID, transfer coretypes.ColoredBalances) bool {
	vmctx.pushCallContext(accountsc.Hname, nil, nil) // create local context for the state
	defer vmctx.popCallContext()

	return accountsc.DebitFromAccount(codec.NewMustCodec(vmctx), agentID, transfer)
}

func (vmctx *VMContext) moveBetweenAccounts(fromAgentID, toAgentID coretypes.AgentID, transfer coretypes.ColoredBalances) bool {
	if len(vmctx.callStack) == 0 {
		vmctx.log.Panicf("moveBetweenAccounts can't be called from request context")
	}

	vmctx.pushCallContext(accountsc.Hname, nil, nil) // create local context for the state
	defer vmctx.popCallContext()

	return accountsc.MoveBetweenAccounts(codec.NewMustCodec(vmctx), fromAgentID, toAgentID, transfer)
}
