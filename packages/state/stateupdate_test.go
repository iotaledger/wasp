package state

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestStateUpdateBasic(t *testing.T) {
	t.Run("default time", func(t *testing.T) {
		su := NewStateUpdate()
		_, ok := su.TimestampMutation()
		require.False(t, ok)
	})
	t.Run("non zero time", func(t *testing.T) {
		nowis := time.Now()
		su := NewStateUpdate(nowis)
		ts, ok := su.TimestampMutation()
		require.True(t, ok)
		require.True(t, nowis.Equal(ts))
	})
	t.Run("serialize zero time", func(t *testing.T) {
		su := NewStateUpdate()
		suBin := su.Bytes()
		su1, err := newStateUpdateFromReader(bytes.NewReader(suBin))
		require.NoError(t, err)
		require.EqualValues(t, suBin, su1.Bytes())
	})
	t.Run("serialize non zero time", func(t *testing.T) {
		nowis := time.Now()
		su := NewStateUpdate(nowis)
		suBin := su.Bytes()
		su1, err := newStateUpdateFromReader(bytes.NewReader(suBin))
		require.NoError(t, err)
		require.EqualValues(t, suBin, su1.Bytes())
		ts, ok := su1.TimestampMutation()
		require.True(t, ok)
		require.True(t, nowis.Equal(ts))
	})
	t.Run("just serialize", func(t *testing.T) {
		nowis := time.Now()
		su := NewStateUpdate(nowis)
		su.Mutations().Set("k", []byte("v"))
		suBin := su.Bytes()
		su1, err := newStateUpdateFromReader(bytes.NewReader(suBin))
		require.NoError(t, err)
		require.EqualValues(t, suBin, su1.Bytes())
		ts, ok := su1.TimestampMutation()
		require.True(t, ok)
		require.True(t, nowis.Equal(ts))
	})
	t.Run("serialize del mutation", func(t *testing.T) {
		nowis := time.Now()
		su := NewStateUpdate(nowis)
		su.Mutations().Set("k", []byte("v"))
		su.Mutations().Del("k")

		su1 := NewStateUpdate(nowis)
		require.NotEqualValues(t, su.Bytes(), su1.Bytes())
	})
	t.Run("state update with block index", func(t *testing.T) {
		su := NewStateUpdateWithBlockIndexMutation(42)
		si, ok := su.StateIndexMutation()
		require.True(t, ok)
		require.EqualValues(t, 42, si)
	})
	t.Run("serialize with block index", func(t *testing.T) {
		su := NewStateUpdateWithBlockIndexMutation(42)
		suBin := su.Bytes()
		su1, err := newStateUpdateFromReader(bytes.NewReader(suBin))
		require.NoError(t, err)
		si, ok := su.StateIndexMutation()
		require.True(t, ok)
		require.EqualValues(t, 42, si)

		si1, ok := su1.StateIndexMutation()
		require.True(t, ok)
		require.EqualValues(t, 42, si1)
	})
}
