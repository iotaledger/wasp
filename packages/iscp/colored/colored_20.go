// the package contain all dependencies with the goshimmmer color model
// only included for IOTA  2.0 ledger
//go:build !l1_15
// +build !l1_15

package colored

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
)

const ColorLength = ledgerstate.ColorLength

var (
	IOTA = ColorFromL1Color(ledgerstate.ColorIOTA)
	MINT = ColorFromL1Color(ledgerstate.ColorMint)
)

func ColorFromL1Color(col ledgerstate.Color) (ret Color) {
	copy(ret[:], col.Bytes())
	return
}

var Balances1IotaL1 = map[ledgerstate.Color]uint64{ledgerstate.ColorIOTA: 1}

// BalancesFromL1Balances creates Balances from ledgerstate.ColoredBalances
func BalancesFromL1Balances(cb *ledgerstate.ColoredBalances) Balances {
	ret := NewBalances()
	if cb != nil {
		cb.ForEach(func(col ledgerstate.Color, balance uint64) bool {
			ret.Set(ColorFromL1Color(col), balance)
			return true
		})
	}
	return ret
}

// BalancesFromL1Map creates Balances from map[ledgerstate.Color]uint64
func BalancesFromL1Map(cb map[ledgerstate.Color]uint64) Balances {
	ret := NewBalances()
	for col, bal := range cb {
		ret.Set(ColorFromL1Color(col), bal)
	}
	return ret
}

func ToL1Map(bals Balances) map[ledgerstate.Color]uint64 {
	ret := make(map[ledgerstate.Color]uint64)
	bals.ForEachRandomly(func(col Color, bal uint64) bool {
		c, _, err := ledgerstate.ColorFromBytes(col[:])
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
			ret.Add(ColorFromL1Color(col), balance)
			return true
		})
	}
	return ret, total
}
