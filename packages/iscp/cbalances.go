// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package iscp

import "github.com/iotaledger/goshimmer/packages/ledgerstate"

// Deprecated:
func EqualColoredBalances(b1, b2 *ledgerstate.ColoredBalances) bool {
	if b1 == b2 {
		return true
	}
	if b1 == nil || b2 == nil {
		return false
	}
	col := make(map[ledgerstate.Color]bool)
	b1.ForEach(func(c ledgerstate.Color, _ uint64) bool {
		col[c] = true
		return true
	})
	b2.ForEach(func(c ledgerstate.Color, _ uint64) bool {
		col[c] = true
		return true
	})
	for c := range col {
		v1, ok1 := b1.Get(c)
		v2, ok2 := b2.Get(c)
		if ok1 != ok2 || v1 != v2 {
			return false
		}
	}
	return true
}

// Deprecated:
func NewTransferIotas(n uint64) *ledgerstate.ColoredBalances {
	return ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{ledgerstate.ColorIOTA: n})
}
