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

func (vctx *sandbox) HarvestFees(amount int64) bool {
	available := vctx.TxBuilder.GetInputBalance(balance.ColorIOTA)
	if available < amount {
		amount = available
	}
	return vctx.TxBuilder.MoveToAddress(vctx.OwnerAddress, balance.ColorIOTA, amount) == nil
}

func (vctx *sandbox) AvailableBalanceFromRequest(col *balance.Color) int64 {
	return vctx.TxBuilder.GetInputBalanceFromTransaction(*col, vctx.RequestRef.Tx.ID())
}

func (vctx *sandbox) MoveTokensFromRequest(targetAddr *address.Address, col *balance.Color, amount int64) bool {
	return vctx.TxBuilder.MoveToAddressFromTransaction(*targetAddr, *col, amount, vctx.RequestRef.Tx.ID()) == nil
}

func (vctx *sandbox) EraseColorFromRequest(targetAddr *address.Address, col *balance.Color, amount int64) bool {
	return vctx.TxBuilder.EraseColorFromTransaction(*targetAddr, *col, amount, vctx.RequestRef.Tx.ID()) == nil
}

func (vctx *sandbox) HarvestFeesFromRequest(amount int64) bool {
	txid := vctx.RequestRef.Tx.ID()
	available := vctx.TxBuilder.GetInputBalanceFromTransaction(balance.ColorIOTA, txid)
	if available < amount {
		amount = available
	}
	return vctx.TxBuilder.MoveToAddressFromTransaction(vctx.OwnerAddress, balance.ColorIOTA, amount, txid) == nil
}
