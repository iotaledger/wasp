// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// An experiment (failed for now) to use simple Shamir Secret Sharing instead of
// a DKG for tests is presented in this file. Not sure, whi it fails.
// Similar setup is working in https://github.com/iotaledger/crypto-tss/blob/main/demo/examples/nonce-sign/main.go

package dss_test

import (
	"testing"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	kyberDSS "go.dedis.ch/kyber/v3/sign/dss"
	"go.dedis.ch/kyber/v3/suites"

	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/gpa/adkg"
	"github.com/iotaledger/wasp/v2/packages/tcrypto"
)

// Just to debug DSS/DKG with a simple Shamir Secret Sharing.
func TestDSS(t *testing.T) {
	n := 4
	f := 1
	suite := tcrypto.DefaultEd25519Suite()
	nodeIDs := gpa.MakeTestNodeIDs(n)
	nodeSKs := map[gpa.NodeID]kyber.Scalar{}
	nodePKs := map[gpa.NodeID]kyber.Point{}
	for i := range nodeIDs {
		nodeSKs[nodeIDs[i]] = suite.Scalar().Pick(suite.RandomStream())
		nodePKs[nodeIDs[i]] = suite.Point().Mul(nodeSKs[nodeIDs[i]], nil)
	}
	_, longPK, long := makeDistKeyShares(suite, nodeIDs, f)

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
		dks[n] = tcrypto.NewDistKeyShare(priShares[i], commits, len(nodeIDs), len(nodeIDs)-f)
	}
	return secKey, pubKey, dks
}
