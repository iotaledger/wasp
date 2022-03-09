package state

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStateUpdateBasic(t *testing.T) {
	t.Run("default time", func(t *testing.T) {
		su := NewStateUpdate()
		_, ok, err := su.timestampMutation()
		require.NoError(t, err)
		require.False(t, ok)
	})
	t.Run("non zero time", func(t *testing.T) {
		currentTime := time.Now()
		su := NewStateUpdate(currentTime)
		ts, ok, err := su.timestampMutation()
		require.NoError(t, err)
		require.True(t, ok)
		require.True(t, currentTime.Equal(ts))
	})
	t.Run("serialize zero time", func(t *testing.T) {
		su := NewStateUpdate()
		suBin := su.Bytes()
		su1, err := newStateUpdateFromReader(bytes.NewReader(suBin))
		require.NoError(t, err)
		require.EqualValues(t, suBin, su1.Bytes())
	})
	t.Run("serialize non zero time", func(t *testing.T) {
		currentTime := time.Now()
		su := NewStateUpdate(currentTime)
		suBin := su.Bytes()
		su1, err := newStateUpdateFromReader(bytes.NewReader(suBin))
		require.NoError(t, err)
		require.EqualValues(t, suBin, su1.Bytes())
		ts, ok, err := su1.timestampMutation()
		require.NoError(t, err)
		require.True(t, ok)
		require.True(t, currentTime.Equal(ts))
	})
	t.Run("just serialize", func(t *testing.T) {
		currentTime := time.Now()
		su := NewStateUpdate(currentTime)
		su.Mutations().Set("k", []byte("v"))
		suBin := su.Bytes()
		su1, err := newStateUpdateFromReader(bytes.NewReader(suBin))
		require.NoError(t, err)
		require.EqualValues(t, suBin, su1.Bytes())
		ts, ok, err := su1.timestampMutation()
		require.NoError(t, err)
		require.True(t, ok)
		require.True(t, currentTime.Equal(ts))
	})
	t.Run("serialize del mutation", func(t *testing.T) {
		currentTime := time.Now()
		su := NewStateUpdate(currentTime)
		su.Mutations().Set("k", []byte("v"))
		su.Mutations().Del("k")

		su1 := NewStateUpdate(currentTime)
		require.NotEqualValues(t, su.Bytes(), su1.Bytes())
	})
}
