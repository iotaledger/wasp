package coretypes

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"io"
)

// ColoredBalances read only
type ColoredBalances interface {
	Balance(color balance.Color) int64
	Iterate(func(color balance.Color, balance int64) bool)
	IterateDeterministic(func(color balance.Color, balance int64) bool)
	Len() uint16
	Equal(b1 ColoredBalances) bool
	AddToMap(m map[balance.Color]int64)
	AsMap() map[balance.Color]int64
	Write(w io.Writer) error
	Read(r io.Reader) error
}
