package util

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/iscp/colored"
)

func OutputBalancesByColor(outputs []ledgerstate.Output) (colored.Balances, uint64) {
	ret := colored.NewBalances()
	total := uint64(0)
	for _, out := range outputs {
		out.Balances().ForEach(func(col ledgerstate.Color, balance uint64) bool {
			total += balance
			ret.Add(colored.Color(col), balance)
			return true
		})
	}
	return ret, total
}
