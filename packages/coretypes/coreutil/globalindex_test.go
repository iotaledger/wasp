package coreutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckpointBasic(t *testing.T) {
	glb := NewGlobalSolidIndex()
	base := glb.GetBaseline()
	require.False(t, base.IsValid())
	base.SetBaseline()
	require.False(t, base.IsValid())
	glb.Set(2)
	base.SetBaseline()
	require.True(t, base.IsValid())
	glb.Invalidate()
	require.False(t, base.IsValid())
	glb.Set(3)
	require.False(t, base.IsValid())
	base.SetBaseline()
	require.True(t, base.IsValid())
}
