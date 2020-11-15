package vmcontext

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
)

func (vmctx *VMContext) Balance(col balance.Color) int64 {
	panic("implement me")
}

func (vmctx *VMContext) MoveBalance(target coretypes.AgentID, col balance.Color, amount int64) bool {
	panic("implement me")
}
