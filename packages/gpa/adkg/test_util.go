// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package adkg

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/sign/dss"
	"go.dedis.ch/kyber/v3/suites"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/gpa/adkg/nonce"
)

// For tests only.
func MakeTestDistributedKey(
	t *testing.T,
	suite suites.Suite,
	nodeIDs []gpa.NodeID,
	nodeSKs map[gpa.NodeID]kyber.Scalar,
	nodePKs map[gpa.NodeID]kyber.Point,
	f int,
	log *logger.Logger,
) (kyber.Point, map[gpa.NodeID]dss.DistKeyShare) {
	n := len(nodeIDs)
	if n == 1 {
		// We don't need to make secret sharing for a single node.
		sk := suite.Scalar().Pick(suite.RandomStream())
		pk := suite.Point().Mul(sk, nil)
		dkss := map[gpa.NodeID]dss.DistKeyShare{
			nodeIDs[0]: &dks{share: &share.PriShare{I: 0, V: sk}, commits: []kyber.Point{pk}},
		}
		return pk, dkss
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
		inputs[nid] = nonce.NewInputStart() // Input is only a signal here.
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
		require.NotNil(t, intermediateOutput)
		require.NotNil(t, intermediateOutput.Indexes)
		require.Len(t, intermediateOutput.Indexes, n-f)
		require.Nil(t, intermediateOutput.PriShare)
		intermediateOutputs[nid] = intermediateOutput
	}
	require.Len(t, intermediateOutputs, n-f)
	//
	// Emulate the agreement.
	decidedProposals := map[gpa.NodeID][]int{}
	for nid := range intermediateOutputs {
		decidedProposals[nid] = intermediateOutputs[nid].Indexes
	}
	//
	// Run the ADKG with agreement already decided.
	for _, nid := range nodeIDs {
		tc.WithInput(nid, nonce.NewInputAgreementResult(decidedProposals))
	}
	tc.WithInputProbability(0.001)
	tc.RunUntil(tc.OutOfMessagesPredicate())
	//
	// Check the FINAL result.
	var pubKey kyber.Point
	dkss := map[gpa.NodeID]dss.DistKeyShare{}
	for nid, n := range nodes {
		o := n.Output()
		require.NotNil(t, o)
		require.NotNil(t, o.(*nonce.Output).PubKey)
		require.NotNil(t, o.(*nonce.Output).PriShare)
		require.NotNil(t, o.(*nonce.Output).Commits)
		dkss[nid] = &dks{share: o.(*nonce.Output).PriShare, commits: o.(*nonce.Output).Commits}
		if pubKey == nil {
			pubKey = o.(*nonce.Output).Commits[0]
		}
	}
	return pubKey, dkss
}

// For tests only.
func VerifyPriShares(
	t *testing.T,
	suite suites.Suite,
	nodeIDs []gpa.NodeID,
	nodePKs map[gpa.NodeID]kyber.Point,
	nodeSKs map[gpa.NodeID]kyber.Scalar,
	longPubKey kyber.Point,
	priShares map[gpa.NodeID]*share.PriShare,
	commits []kyber.Point,
	f int,
) {
	n := len(nodeIDs)
	messageToSign := []byte{112, 117, 116, 105, 110, 32, 99, 104, 117, 105, 108, 111}
	signers := make([]*dss.DSS, n)
	partSigs := make([]*dss.PartialSig, n)
	for i := range nodeIDs {
		nodePKArray := make([]kyber.Point, n)
		for j := range nodePKArray {
			nodePKArray[j] = nodePKs[nodeIDs[j]]
		}
		long := &dks{share: priShares[nodeIDs[i]], commits: commits} // We use long key for nonce as well. Insecure, but OK for this test.
		signer, err := dss.NewDSS(suite, nodeSKs[nodeIDs[i]], nodePKArray, long, long, messageToSign, f+1)
		require.NoError(t, err)
		signers[i] = signer
		partSigs[i], err = signer.PartialSig()
		require.NoError(t, err)
	}
	for i := range nodeIDs {
		for j := range nodeIDs {
			if i == j {
				continue
			}
			if !signers[i].EnoughPartialSig() {
				require.NoError(t, signers[i].ProcessPartialSig(partSigs[j]))
			}
		}
		require.True(t, signers[i].EnoughPartialSig())
		sig, err := signers[i].Signature()
		require.NoError(t, err)
		require.NoError(t, dss.Verify(longPubKey, messageToSign, sig))
	}
}

type dks struct {
	share   *share.PriShare
	commits []kyber.Point
}

func (s *dks) PriShare() *share.PriShare {
	return s.share
}

func (s *dks) Commitments() []kyber.Point {
	return s.commits
}
