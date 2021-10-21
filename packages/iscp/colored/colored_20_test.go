package colored

import (
	"testing"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/stretchr/testify/require"
)

func TestNewColored20Balances(t *testing.T) {
	t.Run("new goshimmer", func(t *testing.T) {
		cb := BalancesFromL1Balances(ledgerstate.NewColoredBalances(nil))
		require.EqualValues(t, 0, len(cb))
	})
}
