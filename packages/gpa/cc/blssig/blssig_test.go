// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package blssig_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/gpa/cc/blssig"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
)

// Here:
//   - HT -- High Threshold (N-F)
//   - LT -- Low Threshold (F+1)
func TestBasic(t *testing.T) {
	t.Run("n=1,t=1", func(tt *testing.T) { testBasic(tt, 1, 10, 0) })
	t.Run("n=2,t=2 (HT)", func(tt *testing.T) { testBasic(tt, 2, 2, 0) })
	t.Run("n=2,t=1 (LT)", func(tt *testing.T) { testBasic(tt, 2, 1, 0) })
	t.Run("n=3,t=3 (HT)", func(tt *testing.T) { testBasic(tt, 3, 3, 0) })
	t.Run("n=3,t=1 (LT)", func(tt *testing.T) { testBasic(tt, 3, 1, 0) })
	t.Run("n=4,t=3 (HT)", func(tt *testing.T) { testBasic(tt, 4, 3, 0) })
	t.Run("n=4,t=2 (LT)", func(tt *testing.T) { testBasic(tt, 4, 2, 0) })
	t.Run("n=10,t=7 (HT)", func(tt *testing.T) { testBasic(tt, 10, 7, 0) })
	t.Run("n=10,t=4 (HT)", func(tt *testing.T) { testBasic(tt, 10, 4, 0) })
}

// All F nodes are silent in this case.
func TestSilent(t *testing.T) {
	t.Run("n=4,t=3 (HT)", func(tt *testing.T) { testBasic(tt, 4, 3, 1) })
	t.Run("n=4,t=2 (LT)", func(tt *testing.T) { testBasic(tt, 4, 2, 1) })
	t.Run("n=10,t=7 (HT)", func(tt *testing.T) { testBasic(tt, 10, 7, 3) })
	t.Run("n=10,t=4 (HT)", func(tt *testing.T) { testBasic(tt, 10, 4, 3) })
}

func testBasic(t *testing.T, nodeCount, threshold, silent int) {
	log := testlogger.NewLogger(t)
	suite := tcrypto.DefaultBLSSuite()
	nodeIDs := gpa.MakeTestNodeIDs(nodeCount)
	nodes := map[gpa.NodeID]gpa.GPA{}
	_, commits, priShares := testpeers.MakeSharedSecret(suite, nodeCount, threshold)
	for i, ni := range nodeIDs {
		if i >= nodeCount-silent {
			nodes[ni] = gpa.MakeTestSilentNode()
		} else {
			nodes[ni] = blssig.New(suite, nodeIDs, commits, priShares[i], threshold, nodeIDs[i], []byte{1, 2, 3}, log)
		}
	}
	inputs := map[gpa.NodeID]gpa.Input{}
	for i := range nodeIDs {
		inputs[nodeIDs[i]] = nil
	}
	tc := gpa.NewTestContext(nodes)
	tc.WithInputs(inputs).RunAll()
	tc.PrintAllStatusStrings("done", t.Logf)
	for i, ni := range nodeIDs {
		if i >= nodeCount-silent {
			continue
		}
		out := nodes[ni].Output()
		require.NotNil(t, out)
		require.Equal(t, nodes[nodeIDs[0]].Output(), out)
	}
}
