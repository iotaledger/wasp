package color

import (
	"bytes"
	"sort"
	"testing"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/stretchr/testify/require"
)

func TestNewColoredBalances(t *testing.T) {
	t.Run("empty 1", func(t *testing.T) {
		var cb Balances
		require.EqualValues(t, 0, len(cb))
		require.NotPanics(t, func() {
			cb.ForEachSorted(func(_ Color, _ uint64) bool {
				return true
			})
		})
		require.EqualValues(t, 0, cb.Get(IOTA))
	})
	t.Run("empty 1", func(t *testing.T) {
		cb := NewBalances(nil)
		require.EqualValues(t, 0, len(cb))
	})
	t.Run("empty 2", func(t *testing.T) {
		cb1 := NewBalances(nil)
		cb2 := NewBalances(nil)
		require.True(t, cb1.Equals(cb2))
	})
	t.Run("empty 3", func(t *testing.T) {
		cb1 := NewBalances(nil)
		cb2 := NewBalances(nil)
		cb2.Set(IOTA, 0)
		require.True(t, cb1.Equals(cb2))
		cb2.Set(IOTA, 5)
		cb2.Set(IOTA, 0)
		require.True(t, cb1.Equals(cb2))
	})
	t.Run("with iotas 1", func(t *testing.T) {
		cb := NewBalances(map[Color]uint64{IOTA: 5})
		require.EqualValues(t, 1, len(cb))
		require.EqualValues(t, 5, cb.Get(IOTA))
	})
	t.Run("with iotas 2", func(t *testing.T) {
		cb := BalancesFromIotas(5)
		require.EqualValues(t, 1, len(cb))
		require.EqualValues(t, 5, cb.Get(IOTA))
	})
	t.Run("with iotas sub", func(t *testing.T) {
		cb := BalancesFromIotas(5)
		require.EqualValues(t, 1, len(cb))
		cb.Sub(IOTA, 10)
		require.EqualValues(t, 0, len(cb))
		require.True(t, cb.IsEmpty())
	})
	t.Run("new goshimmer", func(t *testing.T) {
		cb := BalancesFromLedgerstate1(ledgerstate.NewColoredBalances(nil))
		require.EqualValues(t, 0, len(cb))
	})
	t.Run("equals 1", func(t *testing.T) {
		cb1 := NewBalances(map[Color]uint64{IOTA: 5})
		cb2 := BalancesFromIotas(5)
		require.True(t, cb1.Equals(cb2))
	})
	t.Run("equals 1", func(t *testing.T) {
		cb1 := NewBalances(map[Color]uint64{IOTA: 5})
		cb2 := BalancesFromIotas(5)
		require.True(t, cb1.Equals(cb2))
		cb1.AddAll(cb2)
		require.False(t, cb1.Equals(cb2))
	})
	t.Run("add", func(t *testing.T) {
		cb := BalancesFromIotas(5)
		cb.Add(Mint, 8)
		require.EqualValues(t, 2, len(cb))
	})
	t.Run("marshal1", func(t *testing.T) {
		cb := BalancesFromIotas(5)
		cb.Add(Mint, 8)
		data := cb.Bytes()
		cbBack, err := BalancesFromBytes(data)
		require.NoError(t, err)
		require.True(t, cb.Equals(cbBack))
	})
	t.Run("marshal2", func(t *testing.T) {
		cb := NewBalances(nil)
		data := cb.Bytes()
		cbBack, err := BalancesFromBytes(data)
		require.NoError(t, err)
		require.True(t, cb.Equals(cbBack))
	})
	t.Run("marshal3", func(t *testing.T) {
		var cb Balances
		data := cb.Bytes()
		cbBack, err := BalancesFromBytes(data)
		require.NoError(t, err)
		require.True(t, cb.Equals(cbBack))
	})
	t.Run("for each", func(t *testing.T) {
		const howMany = 100
		arr := make([]Color, howMany)
		cb := NewBalances(nil)
		for i := range arr {
			arr[i] = Random()
			cb.Set(arr[i], uint64(i+1))
		}
		require.EqualValues(t, howMany, len(cb))
		arr1 := make([]Color, howMany)
		idx := 0
		cb.ForEachSorted(func(col Color, _ uint64) bool {
			arr1[idx] = col
			idx++
			return true
		})
		require.True(t, sort.SliceIsSorted(arr1, func(i, j int) bool {
			return bytes.Compare(arr1[i][:], arr1[j][:]) < 0
		}))
	})
}
