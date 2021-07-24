package util

import "github.com/iotaledger/goshimmer/packages/ledgerstate"

func allColors(bals ...*ledgerstate.ColoredBalances) map[ledgerstate.Color]bool {
	ret := make(map[ledgerstate.Color]bool)
	for _, b := range bals {
		b.ForEach(func(col ledgerstate.Color, bal uint64) bool {
			ret[col] = true
			return true
		})
	}
	return ret
}

func DiffColoredBalances(b1, b2 *ledgerstate.ColoredBalances) map[ledgerstate.Color]uint64 {
	ret := make(map[ledgerstate.Color]uint64)
	if b1 == b2 {
		return ret
	}
	for col := range allColors(b1, b2) {
		v1, _ := b1.Get(col)
		v2, _ := b2.Get(col)
		ret[col] = v1 - v2
	}
	return ret
}
