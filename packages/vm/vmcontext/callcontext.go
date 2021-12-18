package vmcontext

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/accounts/commonaccount"
)

// pushCallContextAndMoveAssets moves funds from the caller to the target before pushing new context to the stack
func (vmctx *VMContext) pushCallContextAndMoveAssets(contract iscp.Hname, params dict.Dict, transfer *iscp.Assets) {
	if transfer != nil && commonaccount.IsCoreHname(contract) {
		// if target contract is one of core contracts, transfer is ignored
		transfer = nil
		vmctx.Debugf("transfer ignored for core contract")
	}
	if transfer != nil {
		targetAccount := iscp.NewAgentID(vmctx.ChainID().AsAddress(), contract)
		var sourceAccount *iscp.AgentID
		if len(vmctx.callStack) == 0 {
			sourceAccount = vmctx.req.SenderAccount()
		} else {
			sourceAccount = vmctx.AccountID()
		}
		vmctx.mustMoveBetweenAccounts(sourceAccount, targetAccount, transfer)
	}
	vmctx.pushCallContext(contract, params, transfer)
}

const traceStack = false

func (vmctx *VMContext) pushCallContext(contract iscp.Hname, params dict.Dict, transfer *iscp.Assets) {
	if traceStack {
		vmctx.Debugf("+++++++++++ PUSH %d, stack depth = %d", contract, len(vmctx.callStack))
	}
	var caller *iscp.AgentID
	if len(vmctx.callStack) == 0 {
		// request context
		caller = vmctx.req.SenderAccount()
	} else {
		caller = vmctx.MyAgentID()
	}
	if traceStack {
		vmctx.Debugf("+++++++++++ PUSH %d, stack depth = %d caller = %s", contract, len(vmctx.callStack), caller.String())
	}
	vmctx.callStack = append(vmctx.callStack, &callContext{
		caller:   caller,
		contract: contract,
		params:   params.Clone(),
		transfer: transfer,
	})
}

func (vmctx *VMContext) popCallContext() {
	if traceStack {
		vmctx.Debugf("+++++++++++ POP @ depth %d", len(vmctx.callStack))
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

func (vmctx *VMContext) callCore(c *coreutil.ContractInfo, f func()) {
	vmctx.pushCallContext(c.Hname(), nil, nil)
	defer vmctx.popCallContext()

	f()
}
