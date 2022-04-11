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

func TestBasic(t *testing.T) {
	t.Run("n=4,f=1", func(tt *testing.T) { testBasic(tt, 4, 1) })
	t.Run("n=10,f=3", func(tt *testing.T) { testBasic(tt, 10, 3) })
	t.Run("n=31,f=10", func(tt *testing.T) { testBasic(tt, 31, 10) })
}

func testBasic(t *testing.T, n, f int) {
	nodeIDs := gpa.MakeTestNodeIDs("node", n)
	leader := nodeIDs[rand.Intn(len(nodeIDs))]
	input := []byte("something important to broadcast")
	nodes := map[gpa.NodeID]gpa.GPA{}
	for _, nid := range nodeIDs {
		nodes[nid] = bracha.New(nodeIDs, f, nid, leader, func(b []byte) bool { return true })
	}
	gpa.RunTestWithInputs(nodes, map[gpa.NodeID]gpa.Input{leader: gpa.Input(input)})
	for _, n := range nodes {
		o := n.Output()
		require.NotNil(t, o)
		require.Equal(t, o.([]byte), input)
	}
}
