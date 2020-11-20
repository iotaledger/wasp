// experimental implementation
package accounts

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

func NewColoredBalances() coretypes.ColoredBalances {
	return new(coloredBalances)
}

func NewColoredBalancesFromMap(m map[balance.Color]int64) coretypes.ColoredBalances {
	if m == nil {
		m = make(map[balance.Color]int64)
	}
	return coloredBalances(m)
}

func (b coloredBalances) Balance(col balance.Color) int64 {
	ret, _ := b[col]
	return ret
}

func (b coloredBalances) AsMap() map[balance.Color]int64 {
	return b
}

func (b coloredBalances) String() string {
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

func WriteColoredBalances(w io.Writer, b coretypes.ColoredBalances) error {
	if err := util.WriteUint16(w, b.Len()); err != nil {
		return err
	}
	if b.Len() == 0 {
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
