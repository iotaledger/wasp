// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// package cbalances implements coretypes.ColoredBalances interface
package cbalances

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/util"
	"io"
	"sort"
)

type coloredBalances map[balance.Color]int64

// Nil represents empty colorded balances
var Nil = coretypes.ColoredBalances(coloredBalances(make(map[balance.Color]int64)))

func Str(b coretypes.ColoredBalances) string {
	if b == nil {
		return "[]"
	}
	return b.String()
}

// NewFromMap new ColoredBalances from map
func NewFromMap(m map[balance.Color]int64) coretypes.ColoredBalances {
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

// NewFromBalances from balances in the form of transaction output
func NewFromBalances(bals []*balance.Balance) coretypes.ColoredBalances {
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
	if b == nil || b.Len() == 0 {
		return "(empty)"
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

//
func (b coloredBalances) Len() uint16 {
	return uint16(len(b))
}

func (b coloredBalances) Equal(b1 coretypes.ColoredBalances) bool {
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

func (b coloredBalances) Diff(b1 coretypes.ColoredBalances) coretypes.ColoredBalances {
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

func (b coloredBalances) NonNegative() bool {
	ret := true
	b.Iterate(func(col balance.Color, bal int64) bool {
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

//goland:noinspection ALL
func WriteColoredBalances(w io.Writer, b coretypes.ColoredBalances) error {
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

func ReadColoredBalance(r io.Reader) (coretypes.ColoredBalances, error) {
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
