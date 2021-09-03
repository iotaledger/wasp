// +build !l1_15

package colored

import "github.com/iotaledger/goshimmer/packages/ledgerstate"

// Mint represents a placeholder Color that indicates that tokens should be "colored" in their Output.
var Mint = Color(ledgerstate.ColorMint[:])

var Balances1IotaL1 = ToL1Map(Balances1Iota)

// BalancesFromL1Balances creates Balances from ledgerstate.ColoredBalances
func BalancesFromL1Balances(cb *ledgerstate.ColoredBalances) Balances {
	ret := NewBalances()
	if cb != nil {
		cb.ForEach(func(col ledgerstate.Color, balance uint64) bool {
			ret.Set(col.Bytes(), balance)
			return true
		})
	}
	return ret
}

// BalancesFromL1Map creates Balances from map[ledgerstate.Color]uint64
func BalancesFromL1Map(cb map[ledgerstate.Color]uint64) Balances {
	ret := NewBalances()
	for col, bal := range cb {
		ret.Set(col.Bytes(), bal)
	}
	return ret
}

func ToL1Map(bals Balances) map[ledgerstate.Color]uint64 {
	ret := make(map[ledgerstate.Color]uint64)
	bals.ForEachRandomly(func(col Color, bal uint64) bool {
		c, _, err := ledgerstate.ColorFromBytes(col)
		if err != nil {
			panic(err)
		}
		ret[c] = bal
		return true
	})
	return ret
}

func OutputBalancesByColor(outputs []ledgerstate.Output) (Balances, uint64) {
	ret := NewBalances()
	total := uint64(0)
	for _, out := range outputs {
		out.Balances().ForEach(func(col ledgerstate.Color, balance uint64) bool {
			total += balance
			ret.Add(col.Bytes(), balance)
			return true
		})
	}
	return ret, total
}
