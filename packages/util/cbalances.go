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

func EqualColoredBalances(b1, b2 *ledgerstate.ColoredBalances) bool {
	if b1 == b2 {
		return true
	}
	if b1 == nil || b2 == nil {
		return false
	}
	for col := range allColors(b1, b2) {
		v1, ok1 := b1.Get(col)
		v2, ok2 := b2.Get(col)
		if ok1 != ok2 || v1 != v2 {
			return false
		}
	}
	return true
}

func DiffColoredBalances(b1, b2 *ledgerstate.ColoredBalances) map[ledgerstate.Color]int64 {
	ret := make(map[ledgerstate.Color]int64)
	if b1 == b2 {
		return ret
	}
	for col := range allColors(b1, b2) {
		v1, _ := b1.Get(col)
		v2, _ := b2.Get(col)
		ret[col] = int64(v1) - int64(v2)
	}
	return ret
}
