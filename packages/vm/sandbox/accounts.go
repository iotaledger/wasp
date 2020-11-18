package sandbox

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

func (s *sandbox) Accounts() vmtypes.Accounts {
	return s
}

func (s *sandbox) Incoming() coretypes.ColoredBalances {
	return s.vmctx.GetIncoming()
}

func (s *sandbox) Balance(col balance.Color) int64 {
	return s.vmctx.GetBalance(col)
}

func (s *sandbox) MyBalances() coretypes.ColoredBalances {
	return s.vmctx.GetMyBalances()
}

func (s *sandbox) MoveBalance(target coretypes.AgentID, col balance.Color, amount int64) bool {
	return s.vmctx.DoMoveBalance(target, col, amount)
}
