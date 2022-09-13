// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package nonce_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/gpa/adkg"
	"github.com/iotaledger/wasp/packages/gpa/adkg/nonce"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
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
		tc := gpa.NewTestContext(nodes)
		//
		// Run the DKG
		inputs := make(map[gpa.NodeID]gpa.Input)
		for _, nid := range nodeIDs {
			inputs[nid] = nil // Input is only a signal here.
		}
		tc.WithInputs(inputs).WithInputProbability(0.01)
		tc.RunUntil(tc.NumberOfOutputsPredicate(n - f))
		//
		// Check the INTERMEDIATE result.
		intermediateOutputs := map[gpa.NodeID]*nonce.Output{}
		for nid, node := range nodes {
			nodeOutput := node.Output()
			if nodeOutput == nil {
				continue
			}
			intermediateOutput := nodeOutput.(*nonce.Output)
			require.NotNil(tt, intermediateOutput)
			require.NotNil(tt, intermediateOutput.Indexes)
			require.Len(tt, intermediateOutput.Indexes, n-f)
			require.Nil(tt, intermediateOutput.PriShare)
			intermediateOutputs[nid] = intermediateOutput
		}
		require.Len(tt, intermediateOutputs, n-f)
		//
		// Emulate the agreement.
		decidedProposals := map[gpa.NodeID][]int{}
		for nid := range intermediateOutputs {
			decidedProposals[nid] = intermediateOutputs[nid].Indexes
		}
		//
		// Run the ADKG with agreement already decided.
		agreementMsgs := []gpa.Message{}
		for _, nid := range nodeIDs {
			agreementMsgs = append(agreementMsgs, nonce.NewMsgAgreementResult(nid, decidedProposals))
		}
		tc.WithMessages(agreementMsgs)
		tc.WithInputProbability(0.001)
		tc.RunUntil(tc.OutOfMessagesPredicate())
		//
		// Check the FINAL result.
		priShares := map[gpa.NodeID]*share.PriShare{}
		var pubKey kyber.Point
		var commits []kyber.Point
		for nid, n := range nodes {
			o := n.Output()
			require.NotNil(tt, o)
			require.NotNil(tt, o.(*nonce.Output).PubKey)
			require.NotNil(tt, o.(*nonce.Output).PriShare)
			require.NotNil(tt, o.(*nonce.Output).Commits)
			priShares[nid] = o.(*nonce.Output).PriShare
			if pubKey == nil && commits == nil {
				pubKey = o.(*nonce.Output).PubKey
				commits = o.(*nonce.Output).Commits
			}
		}
		adkg.VerifyPriShares(t, suite, nodeIDs, nodePKs, nodeSKs, pubKey, priShares, commits, f)
	}
	t.Run("n=1,f=0", func(tt *testing.T) { test(tt, 1, 0) })
	t.Run("n=2,f=0", func(tt *testing.T) { test(tt, 2, 0) })
	t.Run("n=3,f=0", func(tt *testing.T) { test(tt, 3, 0) })
	t.Run("n=4,f=1", func(tt *testing.T) { test(tt, 4, 1) })
	t.Run("n=10,f=3", func(tt *testing.T) { test(tt, 10, 3) })
	t.Run("n=31,f=10", func(tt *testing.T) { test(tt, 31, 10) })
}
