package byzquorum_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/util/byzquorum"
)

func TestMaxF(t *testing.T) {
	require.Equal(t, 0, byzquorum.MaxF(0))
	require.Equal(t, 0, byzquorum.MaxF(1))
	require.Equal(t, 0, byzquorum.MaxF(2))
	require.Equal(t, 0, byzquorum.MaxF(3))
	require.Equal(t, 1, byzquorum.MaxF(4))
	require.Equal(t, 1, byzquorum.MaxF(5))
	require.Equal(t, 1, byzquorum.MaxF(6))
	require.Equal(t, 2, byzquorum.MaxF(7))
}

func TestMinQuorum(t *testing.T) {
	require.Equal(t, 0, byzquorum.MinQuorum(0))
	require.Equal(t, 1, byzquorum.MinQuorum(1))
	require.Equal(t, 2, byzquorum.MinQuorum(2))
	require.Equal(t, 3, byzquorum.MinQuorum(3))
	require.Equal(t, 3, byzquorum.MinQuorum(4))
	require.Equal(t, 4, byzquorum.MinQuorum(5))
	require.Equal(t, 5, byzquorum.MinQuorum(6))
	require.Equal(t, 5, byzquorum.MinQuorum(7))
}
