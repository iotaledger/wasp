// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acss_test

import (
	"math/rand"
	"testing"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/gpa/acss"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
)

// In this test all the nodes are actually fair.
func TestBasic(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()
	suite := tcrypto.DefaultEd25519Suite()
	test := func(tt *testing.T, n, f int) {
		nodeIDs := gpa.MakeTestNodeIDs("node", n)
		nodeSKs := map[gpa.NodeID]kyber.Scalar{}
		nodePKs := map[gpa.NodeID]kyber.Point{}
		for i := range nodeIDs {
			nodeSKs[nodeIDs[i]] = suite.Scalar().Pick(suite.RandomStream())
			nodePKs[nodeIDs[i]] = suite.Point().Mul(nodeSKs[nodeIDs[i]], nil)
		}
		dealer := nodeIDs[rand.Intn(len(nodeIDs))]
		nodes := map[gpa.NodeID]gpa.GPA{}
		for _, nid := range nodeIDs {
			nodes[nid] = acss.New(suite, nodeIDs, nodePKs, f, nid, nodeSKs[nid], dealer, log.Named(string(nid)))
		}
		gpa.RunTestWithInputs(nodes, map[gpa.NodeID]gpa.Input{dealer: nil})
		for _, n := range nodes {
			o := n.Output()
			require.NotNil(tt, o)
			// require.Equal(tt, o.([]byte), input) // TODO: Check it.
		}
	}
	t.Parallel()
	t.Run("n=4,f=1", func(tt *testing.T) { test(tt, 4, 1) })
	t.Run("n=10,f=3", func(tt *testing.T) { test(tt, 10, 3) })
	t.Run("n=31,f=10", func(tt *testing.T) { test(tt, 31, 10) })
}
