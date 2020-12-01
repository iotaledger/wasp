package vmcontext

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
)

func (vmctx *VMContext) GetIncoming() coretypes.ColoredBalances {
	return vmctx.getCallContext().transfer
}

func (vmctx *VMContext) GetBalance(col balance.Color) int64 {
	return vmctx.getBalance(col)
}

func (vmctx *VMContext) GetMyBalances() coretypes.ColoredBalances {
	return vmctx.getMyBalances()
}

func (vmctx *VMContext) DoMoveBalance(target coretypes.AgentID, col balance.Color, amount int64) bool {
	return vmctx.moveBalance(target, col, amount)
}
