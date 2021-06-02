package util

import "github.com/iotaledger/goshimmer/packages/ledgerstate"

func OutputBalancesByColor(outputs []ledgerstate.Output) (map[ledgerstate.Color]uint64, uint64) {
	ret := make(map[ledgerstate.Color]uint64)
	total := uint64(0)
	for _, out := range outputs {
		out.Balances().ForEach(func(color ledgerstate.Color, balance uint64) bool {
			total += balance
			ret[color] += balance
			return true
		})
	}
	return ret, total
}
