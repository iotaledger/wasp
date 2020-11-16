// experimental implementation
package accounts

import (
	"bytes"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/util"
	"io"
	"sort"
)

type coloredBalances map[balance.Color]int64

func NewColoredBalances() coretypes.ColoredBalances {
	return new(coloredBalances)
}

func NewColoredBalancesFromMap(m map[balance.Color]int64) coretypes.ColoredBalances {
	return coloredBalances(m)
}

func (b coloredBalances) Balance(col balance.Color) int64 {
	ret, _ := b[col]
	return ret
}

func (b coloredBalances) AsMap() map[balance.Color]int64 {
	return b
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

func (b coloredBalances) Equal(b1 coretypes.ColoredBalances) bool {
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

func (b coloredBalances) AddToMap(m map[balance.Color]int64) {
	b.Iterate(func(col balance.Color, bal int64) bool {
		s, _ := m[col]
		m[col] = s + bal
		return true
	})
}

func (b coloredBalances) Write(w io.Writer) error {
	if err := util.WriteUint16(w, uint16(len(b))); err != nil {
		return err
	}
	if len(b) == 0 {
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

func (b coloredBalances) Read(r io.Reader) error {
	var size uint16
	if err := util.ReadUint16(r, &size); err != nil {
		return err
	}
	if len(b) != 0 {
		// clean if not empty
		t := make([]balance.Color, 0, len(b))
		for k := range b {
			t = append(t, k)
		}
		for _, k := range t {
			delete(b, k)
		}
	}
	if size == 0 {
		return nil
	}
	var col balance.Color
	var bal int64
	for i := uint16(0); i < size; i++ {
		if err := util.ReadColor(r, &col); err != nil {
			return err
		}
		if err := util.ReadInt64(r, &bal); err != nil {
			return err
		}
		b[col] = bal
	}
	return nil
}
