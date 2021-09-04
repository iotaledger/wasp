// the package contain all dependencies with the goshimmmer color model
// only included for IOTA  2.0 ledger
// +build !l1_15

package colored20

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/iscp/colored"
)

func Use() {
	colored.Init(ledgerstate.ColorLength, ColorFromL1Color(ledgerstate.ColorIOTA))
}

var MINT = ColorFromL1Color(ledgerstate.ColorMint)

func ColorFromL1Color(col ledgerstate.Color) colored.Color {
	return col[:]
}

var Balances1IotaL1 = map[ledgerstate.Color]uint64{ledgerstate.ColorIOTA: 1}

// BalancesFromL1Balances creates Balances from ledgerstate.ColoredBalances
func BalancesFromL1Balances(cb *ledgerstate.ColoredBalances) colored.Balances {
	ret := colored.NewBalances()
	if cb != nil {
		cb.ForEach(func(col ledgerstate.Color, balance uint64) bool {
			ret.Set(col.Bytes(), balance)
			return true
		})
	}
	return ret
}

// BalancesFromL1Map creates Balances from map[ledgerstate.Color]uint64
func BalancesFromL1Map(cb map[ledgerstate.Color]uint64) colored.Balances {
	ret := colored.NewBalances()
	for col, bal := range cb {
		ret.Set(col.Bytes(), bal)
	}
	return ret
}

func ToL1Map(bals colored.Balances) map[ledgerstate.Color]uint64 {
	ret := make(map[ledgerstate.Color]uint64)
	bals.ForEachRandomly(func(col colored.Color, bal uint64) bool {
		c, _, err := ledgerstate.ColorFromBytes(col)
		if err != nil {
			panic(err)
		}
		ret[c] = bal
		return true
	})
	return ret
}

func OutputBalancesByColor(outputs []ledgerstate.Output) (colored.Balances, uint64) {
	ret := colored.NewBalances()
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
