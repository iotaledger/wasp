package vmcontext

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coret"
)

func (vmctx *VMContext) GetIncoming() coret.ColoredBalances {
	return vmctx.getCallContext().transfer
}

func (vmctx *VMContext) GetBalance(col balance.Color) int64 {
	return vmctx.getBalance(col)
}

func (vmctx *VMContext) GetMyBalances() coret.ColoredBalances {
	return vmctx.getMyBalances()
}

func (vmctx *VMContext) DoMoveBalance(target coret.AgentID, col balance.Color, amount int64) bool {
	return vmctx.moveBalance(target, col, amount)
}
