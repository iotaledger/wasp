// experimental implementation
package accounts

import (
	"bytes"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"sort"
)

type coloredBalances map[balance.Color]int64

func NewColoredBalancesMutable() coretypes.ColoredBalancesMutable {
	return make(coloredBalances)
}

func (b coloredBalances) Balance(col balance.Color) (int64, bool) {
	ret, ok := b[col]
	return ret, ok
}

func (b coloredBalances) Iterate(f func(col balance.Color, bal int64) bool) {
	for col, bal := range b {
		if !f(col, bal) {
			return
		}
	}
}

func (b coloredBalances) IterateDeterministic(f func(col balance.Color, bal int64) bool) {
	sorted := make([]balance.Color, 0, len(b))
	for col := range b {
		sorted = append(sorted, col)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return bytes.Compare(sorted[i][:], sorted[j][:]) < 0
	})
	for _, col := range sorted {
		if !f(col, b[col]) {
			return
		}
	}
}

func (b coloredBalances) Add(col balance.Color, bal int64) bool {
	v, found := b[col]
	b[col] = v + bal
	return found
}

func (b coloredBalances) Spend(target coretypes.ColoredBalancesMutable, col balance.Color, bal int64) bool {
	v, found := b[col]
	if !found || bal > v {
		return false
	}
	b[col] = v - bal
	if b[col] == 0 {
		delete(b, col)
	}
	target.Add(col, bal)
	return true
}

func (b coloredBalances) SpendAll(target coretypes.ColoredBalancesMutable) {
	b.Iterate(func(col balance.Color, bal int64) bool {
		b.Spend(target, col, bal)
		return true
	})
}
