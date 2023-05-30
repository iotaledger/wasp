package byz_quorum_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/util/byz_quorum"
)

func TestMaxF(t *testing.T) {
	require.Equal(t, 0, byz_quorum.MaxF(0))
	require.Equal(t, 0, byz_quorum.MaxF(1))
	require.Equal(t, 0, byz_quorum.MaxF(2))
	require.Equal(t, 0, byz_quorum.MaxF(3))
	require.Equal(t, 1, byz_quorum.MaxF(4))
	require.Equal(t, 1, byz_quorum.MaxF(5))
	require.Equal(t, 1, byz_quorum.MaxF(6))
	require.Equal(t, 2, byz_quorum.MaxF(7))
}

func TestMinQuorum(t *testing.T) {
	require.Equal(t, 0, byz_quorum.MinQuorum(0))
	require.Equal(t, 1, byz_quorum.MinQuorum(1))
	require.Equal(t, 2, byz_quorum.MinQuorum(2))
	require.Equal(t, 3, byz_quorum.MinQuorum(3))
	require.Equal(t, 3, byz_quorum.MinQuorum(4))
	require.Equal(t, 4, byz_quorum.MinQuorum(5))
	require.Equal(t, 5, byz_quorum.MinQuorum(6))
	require.Equal(t, 5, byz_quorum.MinQuorum(7))
}
