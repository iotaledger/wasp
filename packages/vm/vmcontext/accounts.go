package vmcontext

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/iscp"
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
	return iscp.NewAgentID(vmctx.ChainID().AsAddress(), hname)
}

func (vmctx *VMContext) adjustAccount(agentID *iscp.AgentID) *iscp.AgentID {
	return commonaccount.AdjustIfNeeded(agentID, &vmctx.chainID)
}

func (vmctx *VMContext) commonAccount() *iscp.AgentID {
	return commonaccount.Get(&vmctx.chainID)
}

func (vmctx *VMContext) GetBalance(col ledgerstate.Color) uint64 {
	return vmctx.getBalance(col)
}

func (vmctx *VMContext) GetIncoming() *ledgerstate.ColoredBalances {
	return vmctx.getCallContext().transfer
}

func (vmctx *VMContext) GetMyBalances() *ledgerstate.ColoredBalances {
	return vmctx.getMyBalances()
}
