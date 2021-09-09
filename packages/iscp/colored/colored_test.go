package colored

import (
	"sort"
	"testing"

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
		cb := NewBalances()
		require.EqualValues(t, 0, len(cb))
	})
	t.Run("empty 2", func(t *testing.T) {
		cb1 := NewBalances()
		cb2 := NewBalances()
		require.True(t, cb1.Equals(cb2))
	})
	t.Run("empty 3", func(t *testing.T) {
		cb1 := NewBalances()
		cb2 := NewBalances()
		cb2.Set(IOTA, 0)
		require.True(t, cb1.Equals(cb2))
		cb2.Set(IOTA, 5)
		cb2.Set(IOTA, 0)
		require.True(t, cb1.Equals(cb2))
	})
	t.Run("with iotas 1", func(t *testing.T) {
		cb := NewBalancesForIotas(5)
		require.EqualValues(t, 1, len(cb))
		require.EqualValues(t, 5, cb.Get(IOTA))
	})
	t.Run("with iotas 2", func(t *testing.T) {
		cb := NewBalancesForIotas(5)
		require.EqualValues(t, 1, len(cb))
		require.EqualValues(t, 5, cb.Get(IOTA))
	})
	t.Run("with iotas sub", func(t *testing.T) {
		cb := NewBalancesForIotas(5)
		require.EqualValues(t, 1, len(cb))
		cb.SubNoOverflow(IOTA, 10)
		require.EqualValues(t, 0, len(cb))
		require.True(t, cb.IsEmpty())
	})
	t.Run("equals 1", func(t *testing.T) {
		cb1 := NewBalances()
		cb1.Add(IOTA, 5)
		cb2 := NewBalancesForIotas(5)
		require.True(t, cb1.Equals(cb2))
	})
	t.Run("equals 1", func(t *testing.T) {
		cb1 := NewBalances()
		cb1.Add(IOTA, 5)
		cb2 := NewBalancesForIotas(5)
		require.True(t, cb1.Equals(cb2))
		cb1.AddAll(cb2)
		require.False(t, cb1.Equals(cb2))
	})
	t.Run("add", func(t *testing.T) {
		cb := NewBalancesForIotas(5)
		cb.Add(MINT, 8)
		require.EqualValues(t, 2, len(cb))
	})
	t.Run("marshal1", func(t *testing.T) {
		cb := NewBalancesForIotas(5)
		t.Logf("cb = %s", cb.String())
		cb.Add(MINT, 8)
		data := cb.Bytes()
		t.Logf("cb = %s", cb.String())
		cbBack, err := BalancesFromBytes(data)
		require.NoError(t, err)
		require.True(t, cb.Equals(cbBack))
	})
	t.Run("marshal2", func(t *testing.T) {
		cb := NewBalances()
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
		const howMany = 3
		arr := make([]Color, howMany)
		cb := NewBalances()
		for i := range arr {
			arr[i] = ColorRandom()
			cb.Set(arr[i], uint64(i+1))
		}
		require.EqualValues(t, howMany, len(cb))
		arrToSort := make([]Color, howMany)
		idx := 0
		cb.ForEachSorted(func(col Color, _ uint64) bool {
			arrToSort[idx] = col
			idx++
			return true
		})
		Sort(arrToSort)
		require.True(t, sort.SliceIsSorted(arrToSort, func(i, j int) bool {
			return arrToSort[i].Compare(&arrToSort[j]) < 0
		}))
	})
}

func TestBytesString(t *testing.T) {
	var zeros [32]byte
	z := zeros[:]
	s := string(z)
	t.Logf("s = '%s', len(s) = %d", s, len(s))
	zBack := []byte(s)
	t.Logf("len(zBack) = %d", len(zBack))
	require.EqualValues(t, z, zBack)
}
