// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// These tests are just to avoid "off by 1" mistakes.

package consensus_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/chain/consensus"
)

func TestMaxNodesFaulty(t *testing.T) {
	require.Equal(t, 0, consensus.MaxNodesFaulty(1))
	require.Equal(t, 0, consensus.MaxNodesFaulty(2))
	require.Equal(t, 0, consensus.MaxNodesFaulty(3))
	require.Equal(t, 1, consensus.MaxNodesFaulty(4))
	require.Equal(t, 1, consensus.MaxNodesFaulty(5))
	require.Equal(t, 1, consensus.MaxNodesFaulty(6))
	require.Equal(t, 2, consensus.MaxNodesFaulty(7))
}

func TestMinNodesInQuorum(t *testing.T) {
	require.Equal(t, 1, consensus.MinNodesInQuorum(1))
	require.Equal(t, 2, consensus.MinNodesInQuorum(2))
	require.Equal(t, 3, consensus.MinNodesInQuorum(3))
	require.Equal(t, 3, consensus.MinNodesInQuorum(4))
	require.Equal(t, 4, consensus.MinNodesInQuorum(5))
	require.Equal(t, 5, consensus.MinNodesInQuorum(6))
	require.Equal(t, 5, consensus.MinNodesInQuorum(7))
}

func TestMinNodesFair(t *testing.T) {
	require.Equal(t, 1, consensus.MinNodesFair(1))
	require.Equal(t, 1, consensus.MinNodesFair(2))
	require.Equal(t, 1, consensus.MinNodesFair(3))
	require.Equal(t, 2, consensus.MinNodesFair(4))
	require.Equal(t, 2, consensus.MinNodesFair(5))
	require.Equal(t, 2, consensus.MinNodesFair(6))
	require.Equal(t, 3, consensus.MinNodesFair(7))
}
