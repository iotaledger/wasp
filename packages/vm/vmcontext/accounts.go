package vmcontext

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/accounts/commonaccount"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

func (vmctx *VMContext) AccountID() *iscp.AgentID {
	hname := vmctx.CurrentContractHname()
	switch hname {
	case root.Contract.Hname(), accounts.Contract.Hname(), blob.Contract.Hname(), blocklog.Contract.Hname():
		hname = 0
	}
	return iscp.NewAgentID(vmctx.ChainID().AliasAddress, hname)
}

func (vmctx *VMContext) adjustAccount(agentID *iscp.AgentID) *iscp.AgentID {
	return commonaccount.AdjustIfNeeded(agentID, vmctx.chainID)
}

func (vmctx *VMContext) commonAccount() *iscp.AgentID {
	return commonaccount.Get(vmctx.chainID)
}

func (vmctx *VMContext) GetBalance(col colored.Color) uint64 {
	return vmctx.getBalance(col)
}

func (vmctx *VMContext) GetIncoming() colored.Balances {
	return vmctx.getCallContext().transfer
}

func (vmctx *VMContext) GetMyBalances() colored.Balances {
	return vmctx.getMyBalances()
}
