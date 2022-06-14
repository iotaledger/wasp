// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package adkg

import (
	"testing"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/sign/dss"
	"go.dedis.ch/kyber/v3/suites"
)

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
		long := &dks{share: priShares[nodeIDs[i]], commits: commits}
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
		assert.NoError(t, err)
		assert.NoError(t, dss.Verify(longPubKey, messageToSign, sig))
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
