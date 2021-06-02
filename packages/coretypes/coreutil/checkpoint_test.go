package coreutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckpointBasic(t *testing.T) {
	chp := NewGlobalReadCheckpoint()
	require.False(t, chp.IsValid())
	chp.Start()
	require.True(t, chp.IsValid())
	chp.SetGlobalStateIndex(3)
	require.False(t, chp.IsValid())
	chp.Start()
	require.True(t, chp.IsValid())
}
