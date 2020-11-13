package vmcontext

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
)

func (vmctx *VMContext) AvailableBalance(col *balance.Color) int64 {
	return vmctx.txBuilder.GetInputBalance(*col)
}

func (vmctx *VMContext) MoveTokens(targetAddr *address.Address, col *balance.Color, amount int64) bool {
	return vmctx.txBuilder.MoveTokensToAddress(*targetAddr, *col, amount) == nil
}

func (vmctx *VMContext) EraseColor(targetAddr *address.Address, col *balance.Color, amount int64) bool {
	return vmctx.txBuilder.EraseColor(*targetAddr, *col, amount) == nil
}

func (vmctx *VMContext) AvailableBalanceFromRequest(col *balance.Color) int64 {
	return vmctx.txBuilder.GetInputBalanceFromTransaction(*col, vmctx.reqRef.Tx.ID())
}

func (vmctx *VMContext) MoveTokensFromRequest(targetAddr *address.Address, col *balance.Color, amount int64) bool {
	return vmctx.txBuilder.MoveToAddressFromTransaction(*targetAddr, *col, amount, vmctx.reqRef.Tx.ID()) == nil
}

func (vmctx *VMContext) EraseColorFromRequest(targetAddr *address.Address, col *balance.Color, amount int64) bool {
	return vmctx.txBuilder.EraseColorFromTransaction(*targetAddr, *col, amount, vmctx.reqRef.Tx.ID()) == nil
}
