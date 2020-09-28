package apilib

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
)

func TotalBalanceOfInputs(outs map[valuetransaction.OutputID][]*balance.Balance) int64 {
	ret := int64(0)
	for _, bals := range outs {
		for _, b := range bals {
			ret += b.Value
		}
	}
	return ret
}
