// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// An experiment (failed for now) to use simple Shamir Secret Sharing instead of
// a DKG for tests is presented in this file. Not sure, whi it fails.
// Similar setup is working in https://github.com/iotaledger/crypto-tss/blob/main/demo/examples/nonce-sign/main.go

package dss_test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/chain/dss"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/gpa/adkg"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	kyberDSS "go.dedis.ch/kyber/v3/sign/dss"
	"go.dedis.ch/kyber/v3/suites"
)

// Just to debug DSS/DKG with a simple Shamir Secret Sharing.
func TestDSS(t *testing.T) {
	t.Skipf("why is the simple SSS not working with the DSS?") // TODO: Resolve it somehow.

	n := 4
	f := 1
	suite := tcrypto.DefaultEd25519Suite()
	nodeIDs := gpa.MakeTestNodeIDs("node", n)
	nodeSKs := map[gpa.NodeID]kyber.Scalar{}
	nodePKs := map[gpa.NodeID]kyber.Point{}
	// nodePKArray := make([]kyber.Point, n)
	for i := range nodeIDs {
		nodeSKs[nodeIDs[i]] = suite.Scalar().Pick(suite.RandomStream())
		nodePKs[nodeIDs[i]] = suite.Point().Mul(nodeSKs[nodeIDs[i]], nil)
		// nodePKArray[i] = nodePKs[nodeIDs[i]]
	}
	_, longPK, long := makeDistKeyShares(suite, nodeIDs, f+1)
	// _, _, nonce := makeDistKeyShares(suite, nodeIDs, f+1)

	priShares := make(map[gpa.NodeID]*share.PriShare)
	var commits []kyber.Point
	for _, n := range nodeIDs {
		priShares[n] = long[n].PriShare()
		if commits == nil {
			commits = long[n].Commitments()
		}
	}
	adkg.VerifyPriShares(t, suite, nodeIDs, nodePKs, nodeSKs, longPK, priShares, commits, f)
}

func makeDistKeyShares(suite suites.Suite, nodeIDs []gpa.NodeID, f int) (kyber.Scalar, kyber.Point, map[gpa.NodeID]kyberDSS.DistKeyShare) {
	priPoly := share.NewPriPoly(suite, f+1, nil, suite.RandomStream())
	priShares := priPoly.Shares(len(nodeIDs))
	_, commits := priPoly.Commit(suite.Point().Base()).Info()
	secKey := priPoly.Secret()
	pubKey := commits[0]
	dks := map[gpa.NodeID]kyberDSS.DistKeyShare{}
	for i, n := range nodeIDs {
		dks[n] = dss.NewSecretShare(priShares[i], commits)
	}
	return secKey, pubKey, dks
}
