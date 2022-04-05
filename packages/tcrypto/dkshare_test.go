// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tcrypto

import (
	"testing"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/suites"
	"go.dedis.ch/kyber/v3/util/random"
)

func TestMarshaling(t *testing.T) {
	edSuite, err := suites.Find("Ed25519")
	require.NoError(t, err)
	blsSuite := bn256.NewSuite()
	randomness := random.New()

	nodePubKeys := make([]*cryptolib.PublicKey, 10)
	nodeSecKeys := make([]*cryptolib.PrivateKey, 10)
	for i := range nodePubKeys {
		keyPair := cryptolib.NewKeyPair()
		nodePubKeys[i] = keyPair.GetPublicKey()
		nodeSecKeys[i] = keyPair.GetPrivateKey()
	}

	edPts := make([]kyber.Point, 10)
	for i := range edPts {
		edPts[i] = edSuite.Point().Pick(randomness)
	}

	rnd1 := make([]kyber.Point, 10)
	rnd2 := make([]kyber.Point, 10)
	for i := range rnd1 {
		rnd1[i] = blsSuite.G2().Point().Pick(randomness)
		rnd2[i] = blsSuite.G2().Point().Pick(randomness)
	}

	index := uint16(5)
	dks, err := NewDKShare(
		index,                                   // index
		10,                                      // n
		7,                                       // t
		nodeSecKeys[7],                          // nodePrivKey
		nodePubKeys,                             // nodePubKeys
		edSuite,                                 // edSuite
		edSuite.Point().Pick(randomness),        // edSharedPublic
		edPts,                                   // edPublicCommits
		edPts,                                   // edPublicCommits
		edSuite.Scalar().Pick(randomness),       // edPrivateShare
		blsSuite,                                // blsSuite
		blsSuite.G2().Point().Pick(randomness),  // blsSharedPublic
		rnd1,                                    // blsPublicCommits
		rnd2,                                    // blsPublicShares
		blsSuite.G2().Scalar().Pick(randomness), // blsPrivateShare
	)
	require.NoError(t, err)

	dksBack, err := DKShareFromBytes(dks.Bytes(), edSuite, blsSuite)
	require.NoError(t, err)
	require.EqualValues(t, dks.Bytes(), dksBack.Bytes())
}
