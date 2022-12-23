// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package semi_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/gpa/cc/blssig"
	"github.com/iotaledger/wasp/packages/gpa/cc/semi"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
)

func TestBasic(t *testing.T) {
	t.Run("i=0", func(tt *testing.T) { testBasic(tt, 0) })
	t.Run("i=1", func(tt *testing.T) { testBasic(tt, 1) })
	t.Run("i=2", func(tt *testing.T) { testBasic(tt, 2) })
	t.Run("i=3", func(tt *testing.T) { testBasic(tt, 3) })
	t.Run("i=4", func(tt *testing.T) { testBasic(tt, 4) })
	t.Run("i=5", func(tt *testing.T) { testBasic(tt, 5) })
}

func testBasic(t *testing.T, index int) {
	nodeCount := 4
	threshold := 3
	log := testlogger.NewLogger(t)
	suite := tcrypto.DefaultBLSSuite()
	nodeIDs := gpa.MakeTestNodeIDs("cc", nodeCount)
	nodes := map[gpa.NodeID]gpa.GPA{}
	_, commits, priShares := testpeers.MakeSharedSecret(suite, nodeCount, threshold)
	for i, ni := range nodeIDs {
		target := blssig.New(suite, nodeIDs, commits, priShares[i], threshold, nodeIDs[i], []byte{1, 2, 3}, log)
		nodes[ni] = semi.New(index, target)
	}
	inputs := map[gpa.NodeID]gpa.Input{}
	for i := range nodeIDs {
		inputs[nodeIDs[i]] = nil
	}
	tc := gpa.NewTestContext(nodes)
	tc.WithInputs(inputs).RunAll()
	tc.PrintAllStatusStrings("done", t.Logf)
	for _, ni := range nodeIDs {
		out := nodes[ni].Output()
		require.NotNil(t, out)
		require.Equal(t, nodes[nodeIDs[0]].Output(), out)
	}
}
