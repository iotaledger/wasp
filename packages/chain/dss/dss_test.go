// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dss_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/sign/eddsa"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/chain/dss"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/gpa/adkg"
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

		longTermPK, longTermSecretShares := adkg.MakeTestDistributedKey(tt, suite, nodeIDs, nodeSKs, nodePKs, f, log)

		//
		// Setup nodes.
		dsss := map[gpa.NodeID]dss.DSS{}
		gpas := map[gpa.NodeID]gpa.GPA{}
		for _, nid := range nodeIDs {
			dsss[nid] = dss.New(suite, nodeIDs, nodePKs, f, nid, nodeSKs[nid], longTermSecretShares[nid], log)
			gpas[nid] = dsss[nid].AsGPA()
		}
		tc := gpa.NewTestContext(gpas)
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
		intermediateOutputs := map[gpa.NodeID]*dss.Output{}
		for nid := range gpas {
			nodeOutput := gpas[nid].Output()
			if nodeOutput == nil {
				continue
			}
			intermediateOutput := nodeOutput.(*dss.Output)
			require.NotNil(tt, intermediateOutput)
			require.NotNil(tt, intermediateOutput.ProposedIndexes)
			require.Nil(tt, intermediateOutput.Signature)
			intermediateOutputs[nid] = intermediateOutput
		}
		require.Len(tt, intermediateOutputs, n-f)
		//
		// Emulate the agreement on index proposals (ACS).
		decidedProposals := map[gpa.NodeID][]int{}
		for nid := range intermediateOutputs {
			decidedProposals[nid] = intermediateOutputs[nid].ProposedIndexes
		}
		messageToSign := []byte{112, 117, 116, 105, 110, 32, 99, 104, 117, 105, 108, 111}
		for nid := range dsss {
			tc.WithMessages([]gpa.Message{dsss[nid].NewMsgDecided(decidedProposals, messageToSign)})
		}
		//
		// Run the ADKG with agreement already decided.
		tc.WithInputProbability(0.001)
		tc.RunUntil(tc.OutOfMessagesPredicate())
		//
		// Check the FINAL result.
		var signature []byte
		for _, n := range gpas {
			o := n.Output()
			require.NotNil(tt, o)
			require.NotNil(tt, o.(*dss.Output).Signature)
			if signature == nil {
				signature = o.(*dss.Output).Signature
			}
			require.True(tt, bytes.Equal(signature, o.(*dss.Output).Signature))
		}
		require.NoError(tt, eddsa.Verify(longTermPK, messageToSign, signature))
	}
	t.Run("n=1,f=0", func(tt *testing.T) { test(tt, 1, 0) })
	t.Run("n=2,f=0", func(tt *testing.T) { test(tt, 2, 0) })
	t.Run("n=3,f=0", func(tt *testing.T) { test(tt, 3, 0) })
	t.Run("n=4,f=1", func(tt *testing.T) { test(tt, 4, 1) })
	t.Run("n=10,f=3", func(tt *testing.T) { test(tt, 10, 3) })
	t.Run("n=31,f=10", func(tt *testing.T) { test(tt, 31, 10) })
}
