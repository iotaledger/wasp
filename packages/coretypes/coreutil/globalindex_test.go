package coreutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckpointBasic(t *testing.T) {
	glb := NewChainStateSync()
	base := glb.GetSolidIndexBaseline()
	require.False(t, base.IsValid())
	base.Set()
	require.False(t, base.IsValid())
	glb.SetSolidIndex(2)
	base.Set()
	require.True(t, base.IsValid())
	glb.InvalidateSolidIndex()
	require.False(t, base.IsValid())
	glb.SetSolidIndex(3)
	require.False(t, base.IsValid())
	base.Set()
	require.True(t, base.IsValid())
}
