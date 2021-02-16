package txutil

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
)

func BalancesToString(outs map[valuetransaction.ID][]*balance.Balance) string {
	if outs == nil {
		return "empty balances"
	}

	txids := make([]valuetransaction.ID, 0, len(outs))
	for txid := range outs {
		txids = append(txids, txid)
	}
	sort.Slice(txids, func(i, j int) bool {
		return bytes.Compare(txids[i][:], txids[j][:]) < 0
	})

	ret := ""
	for _, txid := range txids {
		bals := outs[txid]
		ret += txid.String() + ":\n"
		for _, bal := range bals {
			ret += fmt.Sprintf("         %s: %d\n", bal.Color.String(), bal.Value)
		}
	}
	return ret
}

func BalancesByColor(outs map[valuetransaction.ID][]*balance.Balance) (map[balance.Color]int64, int64) {
	ret := make(map[balance.Color]int64)
	var total int64
	for _, bals := range outs {
		for _, b := range bals {
			if s, ok := ret[b.Color]; !ok {
				ret[b.Color] = b.Value
			} else {
				ret[b.Color] = s + b.Value
			}
			total += b.Value
		}
	}
	return ret, total
}

func OutputBalancesByColor(outs map[valuetransaction.OutputID][]*balance.Balance) (map[balance.Color]int64, int64) {
	ret := make(map[balance.Color]int64)
	var total int64
	for _, bals := range outs {
		for _, b := range bals {
			if s, ok := ret[b.Color]; !ok {
				ret[b.Color] = b.Value
			} else {
				ret[b.Color] = s + b.Value
			}
			total += b.Value
		}
	}
	return ret, total
}

func BalanceOfColor(bals []*balance.Balance, color balance.Color) int64 {
	sum := int64(0)
	for _, b := range bals {
		if b.Color == color {
			sum += b.Value
		}
	}
	return sum
}

func CloneBalances(bals []*balance.Balance) []*balance.Balance {
	ret := make([]*balance.Balance, len(bals))
	for i := range ret {
		ret[i] = balance.New(bals[i].Color, bals[i].Value)
	}
	return ret
}

func BalancesSumTotal(bals []*balance.Balance) int64 {
	var ret int64
	for _, b := range bals {
		ret += b.Value
	}
	return ret
}
