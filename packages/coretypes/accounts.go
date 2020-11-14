package coretypes

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
)

// ColoredBalances read only
type ColoredBalances interface {
	Balance(color balance.Color) int64
	Iterate(func(color balance.Color, balance int64) bool)
	IterateDeterministic(func(color balance.Color, balance int64) bool)
}
