// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package nonce_test

import (
	"testing"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/gpa/adkg/nonce"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
)

func TestBasic(t *testing.T) {
	log := testlogger.WithLevel(testlogger.NewLogger(t), logger.LevelWarn, false)
	defer log.Sync()
	suite := tcrypto.DefaultEd25519Suite()
	test := func(tt *testing.T, n, f int) {
		//
		// Setup keys and node names.
		nodeIDs := gpa.MakeTestNodeIDs("node", n)
		nodeSKs := map[gpa.NodeID]kyber.Scalar{}
		nodePKs := map[gpa.NodeID]kyber.Point{}
		for i := range nodeIDs {
			nodeSKs[nodeIDs[i]] = suite.Scalar().Pick(suite.RandomStream())
			nodePKs[nodeIDs[i]] = suite.Point().Mul(nodeSKs[nodeIDs[i]], nil)
		}
		//
		// Setup nodes.
		nodes := map[gpa.NodeID]gpa.GPA{}
		for _, nid := range nodeIDs {
			nodes[nid] = nonce.New(suite, nodeIDs, nodePKs, f, nid, nodeSKs[nid], log)
		}
		//
		// Run the DKG
		inputs := make(map[gpa.NodeID]gpa.Input)
		for _, nid := range nodeIDs {
			inputs[nid] = nil // Input is only a signal here.
		}
		gpa.RunTestWithInputs(nodes, inputs)
		//
		// Check the INTERMEDIATE result.
		for _, node := range nodes {
			o := node.Output()
			require.NotNil(tt, o)
			require.NotNil(tt, o.(*nonce.Output).Indexes)
			require.Len(tt, o.(*nonce.Output).Indexes, n-f)
			require.Nil(tt, o.(*nonce.Output).PriShare)
		}
		//
		// Emulate the agreement (union of index sets from f+1 proposals).
		decidedProposals := gpa.ShuffleNodeIDs(gpa.CopyNodeIDs(nodeIDs))[0 : f+1]
		decidedIndexMap := map[int]bool{}
		for _, nid := range decidedProposals {
			subset := nodes[nid].Output().(*nonce.Output).Indexes
			for _, subsetElem := range subset {
				decidedIndexMap[subsetElem] = true
			}
		}
		decidedIndexes := []int{}
		for idx := range decidedIndexMap {
			decidedIndexes = append(decidedIndexes, idx)
		}
		//
		// Run the ADKG with agreement already decided.
		agreementMsgs := []gpa.Message{}
		for _, nid := range nodeIDs {
			agreementMsgs = append(agreementMsgs, nonce.NewMsgAgreementResult(nid, decidedIndexes))
		}
		gpa.RunTestWithMessages(nodes, agreementMsgs)
		//
		// Check the FINAL result.
		for _, n := range nodes {
			o := n.Output()
			require.NotNil(tt, o)
			require.NotNil(tt, o.(*nonce.Output).PriShare)
		}
	}
	t.Run("n=4,f=1", func(tt *testing.T) { test(tt, 4, 1) })
	t.Run("n=10,f=3", func(tt *testing.T) { test(tt, 10, 3) })
	t.Run("n=31,f=10", func(tt *testing.T) { test(tt, 31, 10) })
}
