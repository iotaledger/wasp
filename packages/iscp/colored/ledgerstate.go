package colored

import "github.com/iotaledger/goshimmer/packages/ledgerstate"

// ColorLength represents the length of a Color (amount of bytes).
const ColorLength = ledgerstate.ColorLength

// IOTA is the zero value of the Color and represents uncolored tokens.
var IOTA = Color(ledgerstate.ColorIOTA)

// Mint represents a placeholder Color that indicates that tokens should be "colored" in their Output.
var Mint = Color(ledgerstate.ColorMint)

var Balances1IotaL1 = ToL1Map(Balances1Iota)

// Color represents a marker that is associated to a token balance and that can give tokens a certain "meaning".
type Color ledgerstate.Color

// BalancesFromL1Balances creates Balances from ledgerstate.ColoredBalances
func BalancesFromL1Balances(cb *ledgerstate.ColoredBalances) Balances {
	ret := NewBalances()
	if cb != nil {
		cb.ForEach(func(color ledgerstate.Color, balance uint64) bool {
			ret.Set(Color(color), balance)
			return true
		})
	}
	return ret
}

// BalancesFromL1Map creates Balances from map[ledgerstate.Color]uint64
func BalancesFromL1Map(cb map[ledgerstate.Color]uint64) Balances {
	ret := NewBalances()
	for col, bal := range cb {
		ret.Set(Color(col), bal)
	}
	return ret
}

func ToL1Map(bals Balances) map[ledgerstate.Color]uint64 {
	ret := make(map[ledgerstate.Color]uint64)
	bals.ForEachRandomly(func(col Color, bal uint64) bool {
		ret[ledgerstate.Color(col)] = bal
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
			ret.Add(Color(col), balance)
			return true
		})
	}
	return ret, total
}
