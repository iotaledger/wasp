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
		require.True(t, su.Timestamp().IsZero())
	})
	t.Run("non zero time", func(t *testing.T) {
		nowis := time.Now()
		su := NewStateUpdate(nowis)
		require.True(t, nowis.Equal(su.Timestamp()))
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
		require.True(t, nowis.Equal(su1.Timestamp()))
	})
	t.Run("just serialize", func(t *testing.T) {
		nowis := time.Now()
		su := NewStateUpdate(nowis)
		su.Mutations().Set("k", []byte("v"))
		suBin := su.Bytes()
		su1, err := newStateUpdateFromReader(bytes.NewReader(suBin))
		require.NoError(t, err)
		require.EqualValues(t, suBin, su1.Bytes())
		require.True(t, nowis.Equal(su1.Timestamp()))
	})
	t.Run("serialize del mutation", func(t *testing.T) {
		nowis := time.Now()
		su := NewStateUpdate(nowis)
		su.Mutations().Set("k", []byte("v"))
		su.Mutations().Del("k")

		su1 := NewStateUpdate(nowis)
		require.NotEqualValues(t, su.Bytes(), su1.Bytes())
	})
}
