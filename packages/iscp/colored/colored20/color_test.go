package colored20

import (
	"bytes"
	"sort"
	"testing"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/stretchr/testify/require"
)

func TestNewColoredBalances(t *testing.T) {
	Use()

	t.Run("empty 1", func(t *testing.T) {
		var cb colored.Balances
		require.EqualValues(t, 0, len(cb))
		require.NotPanics(t, func() {
			cb.ForEachSorted(func(_ colored.Color, _ uint64) bool {
				return true
			})
		})
		require.EqualValues(t, 0, cb.Get(colored.IOTA))
	})
	t.Run("empty 1", func(t *testing.T) {
		cb := colored.NewBalances()
		require.EqualValues(t, 0, len(cb))
	})
	t.Run("empty 2", func(t *testing.T) {
		cb1 := colored.NewBalances()
		cb2 := colored.NewBalances()
		require.True(t, cb1.Equals(cb2))
	})
	t.Run("empty 3", func(t *testing.T) {
		cb1 := colored.NewBalances()
		cb2 := colored.NewBalances()
		cb2.Set(colored.IOTA, 0)
		require.True(t, cb1.Equals(cb2))
		cb2.Set(colored.IOTA, 5)
		cb2.Set(colored.IOTA, 0)
		require.True(t, cb1.Equals(cb2))
	})
	t.Run("with iotas 1", func(t *testing.T) {
		cb := colored.NewBalancesForIotas(5)
		require.EqualValues(t, 1, len(cb))
		require.EqualValues(t, 5, cb.Get(colored.IOTA))
	})
	t.Run("with iotas 2", func(t *testing.T) {
		cb := colored.NewBalancesForIotas(5)
		require.EqualValues(t, 1, len(cb))
		require.EqualValues(t, 5, cb.Get(colored.IOTA))
	})
	t.Run("with iotas sub", func(t *testing.T) {
		cb := colored.NewBalancesForIotas(5)
		require.EqualValues(t, 1, len(cb))
		cb.SubNoOverflow(colored.IOTA, 10)
		require.EqualValues(t, 0, len(cb))
		require.True(t, cb.IsEmpty())
	})
	t.Run("new goshimmer", func(t *testing.T) {
		cb := BalancesFromL1Balances(ledgerstate.NewColoredBalances(nil))
		require.EqualValues(t, 0, len(cb))
	})
	t.Run("equals 1", func(t *testing.T) {
		cb1 := colored.NewBalances()
		cb1.Add(colored.IOTA, 5)
		cb2 := colored.NewBalancesForIotas(5)
		require.True(t, cb1.Equals(cb2))
	})
	t.Run("equals 1", func(t *testing.T) {
		cb1 := colored.NewBalances()
		cb1.Add(colored.IOTA, 5)
		cb2 := colored.NewBalancesForIotas(5)
		require.True(t, cb1.Equals(cb2))
		cb1.AddAll(cb2)
		require.False(t, cb1.Equals(cb2))
	})
	t.Run("add", func(t *testing.T) {
		cb := colored.NewBalancesForIotas(5)
		cb.Add(MINT, 8)
		require.EqualValues(t, 2, len(cb))
	})
	t.Run("marshal1", func(t *testing.T) {
		cb := colored.NewBalancesForIotas(5)
		t.Logf("cb = %s", cb.String())
		cb.Add(MINT, 8)
		data := cb.Bytes()
		t.Logf("cb = %s", cb.String())
		cbBack, err := colored.BalancesFromBytes(data)
		require.NoError(t, err)
		require.True(t, cb.Equals(cbBack))
	})
	t.Run("marshal2", func(t *testing.T) {
		cb := colored.NewBalances()
		data := cb.Bytes()
		cbBack, err := colored.BalancesFromBytes(data)
		require.NoError(t, err)
		require.True(t, cb.Equals(cbBack))
	})
	t.Run("marshal3", func(t *testing.T) {
		var cb colored.Balances
		data := cb.Bytes()
		cbBack, err := colored.BalancesFromBytes(data)
		require.NoError(t, err)
		require.True(t, cb.Equals(cbBack))
	})
	t.Run("for each", func(t *testing.T) {
		const howMany = 100
		arr := make([]colored.Color, howMany)
		cb := colored.NewBalances()
		for i := range arr {
			arr[i] = colored.ColorRandom()
			cb.Set(arr[i], uint64(i+1))
		}
		require.EqualValues(t, howMany, len(cb))
		arr1 := make([]colored.Color, howMany)
		idx := 0
		cb.ForEachSorted(func(col colored.Color, _ uint64) bool {
			arr1[idx] = col
			idx++
			return true
		})
		require.True(t, sort.SliceIsSorted(arr1, func(i, j int) bool {
			return bytes.Compare(arr1[i][:], arr1[j][:]) < 0
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
