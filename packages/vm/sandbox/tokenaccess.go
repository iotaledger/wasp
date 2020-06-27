package sandbox

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
)

func (vctx *sandbox) AvailableBalance(col *balance.Color) int64 {
	return vctx.TxBuilder.GetInputBalance(*col)
}

func (vctx *sandbox) MoveTokens(targetAddr *address.Address, col *balance.Color, amount int64) bool {
	return vctx.TxBuilder.MoveToAddress(*targetAddr, *col, amount) == nil
}

func (vctx *sandbox) EraseColor(targetAddr *address.Address, col *balance.Color, amount int64) bool {
	return vctx.TxBuilder.EraseColor(*targetAddr, *col, amount) == nil
}
