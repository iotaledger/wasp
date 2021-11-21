package vmcontext

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/vm/core/accounts/commonaccount"
)

func (vmctx *VMContext) AccountID() *iscp.AgentID {
	hname := vmctx.CurrentContractHname()
	if commonaccount.IsCoreHname(hname) {
		return vmctx.commonAccount()
	}
	return iscp.NewAgentID(vmctx.task.AnchorOutput.AliasID.ToAddress(), hname)
}

func (vmctx *VMContext) adjustAccount(agentID *iscp.AgentID) *iscp.AgentID {
	return commonaccount.AdjustIfNeeded(agentID, vmctx.ChainID())
}

func (vmctx *VMContext) commonAccount() *iscp.AgentID {
	return commonaccount.Get(vmctx.ChainID())
}

// Deprecated:
func (vmctx *VMContext) GetBalanceOld(col colored.Color) uint64 {
	panic("deprecated")
}

func (vmctx *VMContext) IncomingTransfer() *iscp.Assets {
	return vmctx.getCallContext().transfer
}
