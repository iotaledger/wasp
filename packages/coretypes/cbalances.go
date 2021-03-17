// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package coretypes

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/util"
	"io"
)

// ColoredBalances is a wrapper of ledgerstate.ColoredBalances
type ColoredBalances struct {
	ledgerstate.ColoredBalances
}

func NewColoredBalances(b ledgerstate.ColoredBalances) ColoredBalances {
	return ColoredBalances{ColoredBalances: *b.Clone()}
}

// NewColoredBalancesFromMap new ColoredBalancesOld from map
func NewColoredBalancesFromMap(m map[ledgerstate.Color]uint64) ColoredBalances {
	return ColoredBalances{*ledgerstate.NewColoredBalances(m)}
}

func NewIotasOnly(amount uint64) ColoredBalances {
	return NewColoredBalancesFromMap(map[ledgerstate.Color]uint64{ledgerstate.ColorIOTA: amount})
}

func (b *ColoredBalances) Clone() *ColoredBalances {
	return &ColoredBalances{*b.ColoredBalances.Clone()}
}

func (b *ColoredBalances) Balance(col ledgerstate.Color) uint64 {
	ret, _ := b.Get(col)
	return ret
}

func (b *ColoredBalances) Len() uint16 {
	return uint16(b.Size())
}

func (b *ColoredBalances) Equal(b1 ColoredBalances) bool {
	ret := true
	b.ForEach(func(col ledgerstate.Color, bal uint64) bool {
		if bal != b1.Balance(col) {
			ret = false
			return false
		}
		return true
	})
	return ret
}

func (b *ColoredBalances) Diff(b1 ColoredBalances) map[ledgerstate.Color]int64 {
	ret := make(map[ledgerstate.Color]int64)
	allColors := make(map[ledgerstate.Color]bool)
	b.ForEach(func(col ledgerstate.Color, _ uint64) bool {
		allColors[col] = true
		return true
	})
	b1.ForEach(func(col ledgerstate.Color, _ uint64) bool {
		allColors[col] = true
		return true
	})
	for col := range allColors {
		ret[col] = int64(b.Balance(col)) - int64(b1.Balance(col))
	}
	return ret
}

func (b *ColoredBalances) AddToMap(m map[ledgerstate.Color]int64) {
	b.ForEach(func(col ledgerstate.Color, bal uint64) bool {
		s, _ := m[col]
		m[col] = s + int64(bal)
		return true
	})
}

func NonNegative(m map[ledgerstate.Color]int64) bool {
	for _, b := range m {
		if b < 0 {
			return false
		}
	}
	return true
}

func MustToUint64(m map[ledgerstate.Color]int64) map[ledgerstate.Color]uint64 {
	ret := make(map[ledgerstate.Color]uint64)
	for k, v := range m {
		if v < 0 {
			panic("MustToUint64: must be non-negative")
		}
		ret[k] = uint64(v)
	}
	return ret
}

// TakeColor takes out all tokens with specific color
// return what has left
func (b *ColoredBalances) TakeOutColor(col ledgerstate.Color) ColoredBalances {
	m := b.Map()
	delete(m, col)
	return NewColoredBalancesFromMap(m)
}

func (b *ColoredBalances) AboveDustThreshold(dustThreshold map[ledgerstate.Color]uint64) bool {
	if len(dustThreshold) == 0 {
		return true
	}
	if b == nil {
		return false
	}
	m := b.ColoredBalances.Map()
	for col, dust := range dustThreshold {
		b, _ := m[col]
		if b < dust {
			return false
		}
	}
	return true
}

func WriteColoredBalances(w io.Writer, b *ColoredBalances) error {
	return util.WriteBytes16(w, b.Bytes())
}

func ReadColoredBalance(r io.Reader, cb *ColoredBalances) error {
	data, err := util.ReadBytes16(r)
	if err != nil {
		return err
	}
	res, _, err := ledgerstate.ColoredBalancesFromBytes(data)
	if err != nil {
		return err
	}
	cb.ColoredBalances = *res
	return nil
}
