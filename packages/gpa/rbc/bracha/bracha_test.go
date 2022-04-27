// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package bracha_test

import (
	"math/rand"
	"testing"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/gpa/rbc/bracha"
	"github.com/stretchr/testify/require"
)

// In this test all the nodes are actually fair.
func TestBasic(t *testing.T) {
	test := func(tt *testing.T, n, f int) {
		nodeIDs := gpa.MakeTestNodeIDs("node", n)
		leader := nodeIDs[rand.Intn(len(nodeIDs))]
		input := []byte("something important to broadcast")
		nodes := map[gpa.NodeID]gpa.GPA{}
		for _, nid := range nodeIDs {
			nodes[nid] = bracha.New(nodeIDs, f, nid, leader, func(b []byte) bool { return true })
		}
		gpa.NewTestContext(nodes).WithInputs(map[gpa.NodeID]gpa.Input{leader: gpa.Input(input)}).RunAll()
		for _, n := range nodes {
			o := n.Output()
			require.NotNil(tt, o)
			require.Equal(tt, o.([]byte), input)
		}
	}
	t.Parallel()
	t.Run("n=4,f=1", func(tt *testing.T) { test(tt, 4, 1) })
	t.Run("n=10,f=3", func(tt *testing.T) { test(tt, 10, 3) })
	t.Run("n=31,f=10", func(tt *testing.T) { test(tt, 31, 10) })
}

// Assume f nodes are actually faulty by dropping all the messages.
func TestWithSilent(t *testing.T) {
	test := func(tt *testing.T, n, f int) {
		nodeIDs := gpa.ShuffleNodeIDs(gpa.MakeTestNodeIDs("node", n))
		faulty := nodeIDs[0:f]
		fair := nodeIDs[f:]
		require.Len(t, faulty, f)
		require.Len(t, fair, n-f)
		leader := fair[0]
		input := []byte("something important to broadcast")
		nodes := map[gpa.NodeID]gpa.GPA{}
		for _, nid := range fair {
			nodes[nid] = bracha.New(nodeIDs, f, nid, leader, func(b []byte) bool { return true })
		}
		for _, nid := range faulty {
			nodes[nid] = gpa.MakeTestSilentNode()
		}
		gpa.NewTestContext(nodes).WithInputs(map[gpa.NodeID]gpa.Input{leader: gpa.Input(input)}).RunAll()
		for _, nid := range fair {
			o := nodes[nid].Output()
			require.NotNil(tt, o)
			require.Equal(tt, o.([]byte), input)
		}
	}
	t.Parallel()
	t.Run("n=4,f=1", func(tt *testing.T) { test(tt, 4, 1) })
	t.Run("n=10,f=3", func(tt *testing.T) { test(tt, 10, 3) })
	t.Run("n=31,f=10", func(tt *testing.T) { test(tt, 31, 10) })
}

// Check if predicate is considered properly.
func TestPredicate(t *testing.T) {
	pTrue := func(b []byte) bool { return true }
	pFalse := func(b []byte) bool { return false }
	test := func(tt *testing.T, n, f int) {
		nodeIDs := gpa.MakeTestNodeIDs("node", n)
		leader := nodeIDs[rand.Intn(len(nodeIDs))]
		input := []byte("something important to broadcast")
		nodes := map[gpa.NodeID]gpa.GPA{}
		for _, nid := range nodeIDs {
			nodes[nid] = bracha.New(nodeIDs, f, nid, leader, pFalse) // NOTE: Initially false.
		}
		//
		// No outputs are returned while predicates are false.
		tc := gpa.NewTestContext(nodes).WithInputs(map[gpa.NodeID]gpa.Input{leader: gpa.Input(input)})
		tc.RunAll()
		for nid := range nodes {
			require.Nil(t, nodes[nid].Output())
		}
		//
		// Update the predicate to TRUE for at least N-F nodes.
		// That should make the RBC to terminate at all the fair instances.
		predicateUpdates := make([]gpa.Message, n-f)
		for i := range predicateUpdates {
			predicateUpdates[i] = bracha.MakePredicateUpdateMsg(nodeIDs[i], pTrue)
		}
		tc.WithMessages(predicateUpdates).RunAll()
		for nid := range nodes {
			o := nodes[nid].Output()
			require.NotNil(tt, o)
			require.Equal(tt, o.([]byte), input)
		}
	}
	t.Parallel()
	t.Run("n=4,f=1", func(tt *testing.T) { test(tt, 4, 1) })
	t.Run("n=10,f=3", func(tt *testing.T) { test(tt, 10, 3) })
	t.Run("n=31,f=10", func(tt *testing.T) { test(tt, 31, 10) })
}
