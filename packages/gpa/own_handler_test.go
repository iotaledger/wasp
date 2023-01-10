// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRound uses OwnHandler, so we use it for this test.
func TestOwnHandler(t *testing.T) {
	t.Parallel()
	n := 10
	nodeIDs := MakeTestNodeIDs(n)
	nodes := map[NodeID]GPA{}
	inputs := map[NodeID]Input{}
	for _, nid := range nodeIDs {
		nodes[nid] = NewTestRound(nodeIDs, nid)
		inputs[nid] = nil
	}
	tc := NewTestContext(nodes).WithInputs(inputs).WithInputProbability(0.5)
	tc.RunAll()
	for _, n := range nodes {
		require.NotNil(t, n.Output())
	}
}
