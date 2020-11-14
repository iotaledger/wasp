// experimental implementation
package accounts

import (
	"bytes"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"sort"
)

type coloredBalances map[balance.Color]int64

func FromMap(m map[balance.Color]int64) coretypes.ColoredBalances {
	return coloredBalances(m)
}

func (b coloredBalances) Balance(col balance.Color) int64 {
	ret, _ := b[col]
	return ret
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
