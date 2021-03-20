package vmcontext

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
)

func (vmctx *VMContext) GetIncoming() *ledgerstate.ColoredBalances {
	return vmctx.getCallContext().transfer
}

func (vmctx *VMContext) GetBalance(col ledgerstate.Color) uint64 {
	return vmctx.getBalance(col)
}

func (vmctx *VMContext) GetMyBalances() *ledgerstate.ColoredBalances {
	return vmctx.getMyBalances()
}

func (vmctx *VMContext) commonAccount() *coretypes.AgentID {
	return coretypes.NewAgentID(vmctx.chainID.AsAddress(), 0)
}

func (vmctx *VMContext) adjustAccount(agentID *coretypes.AgentID) *coretypes.AgentID {
	if !IsAdjustableAccount(vmctx.chainID, &vmctx.chainOwnerID, agentID) {
		return agentID
	}
	return vmctx.commonAccount()
}
