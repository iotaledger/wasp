package vmcontext

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/accounts/commonaccount"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/eventlog"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

func (vmctx *VMContext) AccountID() *coretypes.AgentID {
	hname := vmctx.CurrentContractHname()
	switch hname {
	case root.Interface.Hname(), accounts.Interface.Hname(), blob.Interface.Hname(), eventlog.Interface.Hname():
		hname = 0
	}
	return coretypes.NewAgentID(vmctx.ChainID().AsAddress(), hname)
}

func (vmctx *VMContext) adjustAccount(agentID *coretypes.AgentID) *coretypes.AgentID {
	return commonaccount.AdjustIfNeeded(agentID, &vmctx.chainID)
}

func (vmctx *VMContext) commonAccount() *coretypes.AgentID {
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
