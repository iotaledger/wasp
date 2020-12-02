// implements coret.ColoredBalances interface
package cbalances

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coret"
	"github.com/iotaledger/wasp/packages/util"
	"io"
	"sort"
)

type coloredBalances map[balance.Color]int64

var Nil = coret.ColoredBalances(coloredBalances(make(map[balance.Color]int64)))

func Str(b coret.ColoredBalances) string {
	if b == nil {
		return "[]"
	}
	return b.String()
}

func NewFromMap(m map[balance.Color]int64) coret.ColoredBalances {
	if m == nil {
		return Nil
	}
	ret := make(map[balance.Color]int64)
	for c, b := range m {
		if b != 0 {
			ret[c] = b
		}
	}
	return coloredBalances(ret)
}

func NewFromBalances(bals []*balance.Balance) coret.ColoredBalances {
	if bals == nil {
		return Nil
	}
	ret := make(map[balance.Color]int64, len(bals))
	for _, b := range bals {
		s, _ := ret[b.Color]
		ret[b.Color] = s + b.Value
	}
	return NewFromMap(ret)
}

func (b coloredBalances) Balance(col balance.Color) int64 {
	if b == nil {
		return 0
	}
	ret, _ := b[col]
	return ret
}

func (b coloredBalances) String() string {
	if b == nil {
		return ""
	}
	ret := ""
	b.IterateDeterministic(func(col balance.Color, bal int64) bool {
		ret += fmt.Sprintf("       %s: %d\n", col.String(), bal)
		return true
	})
	return ret
}

func (b coloredBalances) Iterate(f func(col balance.Color, bal int64) bool) {
	for col, bal := range b {
		if bal == 0 {
			continue
		}
		if !f(col, bal) {
			return
		}
	}
}

func (b coloredBalances) IterateDeterministic(f func(col balance.Color, bal int64) bool) {
	sorted := make([]balance.Color, 0, len(b))
	for col, bal := range b {
		if bal == 0 {
			continue
		}
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

func (b coloredBalances) Len() uint16 {
	return uint16(len(b))
}

func (b coloredBalances) Equal(b1 coret.ColoredBalances) bool {
	if b == nil && b1 == nil {
		return true
	}
	if (b == nil) != (b1 == nil) {
		return false
	}
	if b.Len() != b1.Len() {
		return false
	}
	ret := true
	b.Iterate(func(col balance.Color, bal int64) bool {
		if bal != b1.Balance(col) {
			ret = false
			return false
		}
		return true
	})
	return ret
}

// Diff return difference between the two: b - b1
func (b coloredBalances) Diff(b1 coret.ColoredBalances) coret.ColoredBalances {
	ret := make(map[balance.Color]int64)
	if b == nil && b1 == nil {
		return Nil
	}
	allColors := make(map[balance.Color]bool)
	for c := range b {
		allColors[c] = true
	}
	b1.Iterate(func(c balance.Color, _ int64) bool {
		allColors[c] = true
		return true
	})
	for col := range allColors {
		s, _ := b[col]
		s1 := b1.Balance(col)
		if s != s1 {
			ret[col] = s - s1
		}
	}
	return NewFromMap(ret)
}

// Includes b >= b1
func (b coloredBalances) Includes(b1 coret.ColoredBalances) bool {
	diff := b.Diff(b1)
	if diff == nil || diff.Len() == 0 {
		return true
	}
	ret := true
	diff.Iterate(func(col balance.Color, bal int64) bool {
		if bal < 0 {
			ret = false
			return false
		}
		return true
	})
	return ret
}

func (b coloredBalances) AddToMap(m map[balance.Color]int64) {
	b.Iterate(func(col balance.Color, bal int64) bool {
		s, _ := m[col]
		m[col] = s + bal
		return true
	})
}

func WriteColoredBalances(w io.Writer, b coret.ColoredBalances) error {
	l := uint16(0)
	if b != nil {
		l = b.Len()
	}
	if err := util.WriteUint16(w, l); err != nil {
		return err
	}
	if l == 0 {
		return nil
	}
	var err error
	b.IterateDeterministic(func(col balance.Color, bal int64) bool {
		if _, e := w.Write(col[:]); e != nil {
			err = e
			return false
		}
		if e := util.WriteInt64(w, bal); e != nil {
			err = e
			return false
		}
		return true
	})
	return err
}

func ReadColoredBalance(r io.Reader) (coret.ColoredBalances, error) {
	var size uint16
	if err := util.ReadUint16(r, &size); err != nil {
		return nil, err
	}
	ret := make(coloredBalances)
	var col balance.Color
	var bal int64
	for i := uint16(0); i < size; i++ {
		if err := util.ReadColor(r, &col); err != nil {
			return nil, err
		}
		if err := util.ReadInt64(r, &bal); err != nil {
			return nil, err
		}
		ret[col] = bal
	}
	return ret, nil
}
