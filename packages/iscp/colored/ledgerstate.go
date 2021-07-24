package colored

import "github.com/iotaledger/goshimmer/packages/ledgerstate"

// ColorLength represents the length of a Color (amount of bytes).
const ColorLength = ledgerstate.ColorLength

// Color represents a marker that is associated to a token balance and that can give tokens a certain "meaning".
type Color ledgerstate.Color

func BalancesFromLedgerstate1(cb *ledgerstate.ColoredBalances) Balances {
	ret := NewBalances()
	if cb != nil {
		cb.ForEach(func(color ledgerstate.Color, balance uint64) bool {
			ret.Set(Color(color), balance)
			return true
		})
	}
	return ret
}

func BalancesFromLedgerstate2(cb map[ledgerstate.Color]uint64) Balances {
	ret := NewBalances()
	for col, bal := range cb {
		ret.Set(Color(col), bal)
	}
	return ret
}

func ToLedgerstateMap(bals Balances) map[ledgerstate.Color]uint64 {
	ret := make(map[ledgerstate.Color]uint64)
	bals.ForEachRandomly(func(col Color, bal uint64) bool {
		ret[ledgerstate.Color(col)] = bal
		return true
	})
	return ret
}
