// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testpeers

import (
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"fortio.org/safecast"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/suites"

	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/distkeygen"
	"github.com/iotaledger/wasp/v2/packages/peering"
	"github.com/iotaledger/wasp/v2/packages/registry"
	"github.com/iotaledger/wasp/v2/packages/tcrypto"
	"github.com/iotaledger/wasp/v2/packages/testutil"
	"github.com/iotaledger/wasp/v2/packages/testutil/testlogger"
)

func SetupKeys(peerCount uint16) ([]string, []*cryptolib.KeyPair) {
	peeringURLs := make([]string, peerCount)
	peerIdentities := make([]*cryptolib.KeyPair, peerCount)
	for i := range peeringURLs {
		peerIdentities[i] = cryptolib.NewKeyPair()
		peeringURLs[i] = fmt.Sprintf("P%02d", i)
	}
	return peeringURLs, peerIdentities
}

func PublicKeys(peerIdentities []*cryptolib.KeyPair) []*cryptolib.PublicKey {
	pubKeys := make([]*cryptolib.PublicKey, len(peerIdentities))
	for i := range pubKeys {
		pubKeys[i] = peerIdentities[i].GetPublicKey()
	}
	return pubKeys
}

func SetupDistributedKeyGeneration(
	t *testing.T,
	threshold uint16,
	peeringURLs []string,
	peerIdentities []*cryptolib.KeyPair,
	suite tcrypto.Suite,
	log log.Logger,
) (*cryptolib.Address, []registry.DKShareRegistry) {
	timeout := 300 * time.Second
	networkProviders, networkCloser := SetupNet(peeringURLs, peerIdentities, testutil.NewPeeringNetReliable(log), log)
	//
	// Initialize the DKG subsystem in each node.
	distKeyGenNodes := make([]*distkeygen.Node, len(peeringURLs))
	dkShareRegistrys := make([]registry.DKShareRegistry, len(peeringURLs))
	for i := range peeringURLs {
		dkShareRegistrys[i] = testutil.NewDistributedKeyGenerationRegistry(peerIdentities[i].GetPrivateKey())
		distKeyGenNode, err := distkeygen.NewNode(
			peerIdentities[i], networkProviders[i], dkShareRegistrys[i],
			testlogger.WithLevel(log.NewChildLogger(fmt.Sprintf("peeringURL:%s", peeringURLs[i])), slog.LevelError, false),
		)
		require.NoError(t, err)
		distKeyGenNodes[i] = distKeyGenNode
	}
	//
	// Initiate the key generation from some client node.
	dkShare, err := distKeyGenNodes[0].GenerateDistributedKey(
		PublicKeys(peerIdentities),
		threshold,
		100*time.Second,
		200*time.Second,
		timeout,
	)
	require.NoError(t, err)
	require.NotNil(t, dkShare.GetAddress())
	require.NotNil(t, dkShare.GetSharedPublic())
	require.NoError(t, networkCloser.Close())
	return dkShare.GetAddress(), dkShareRegistrys
}

func SetupDistributedKeyGenerationTrivial(
	t require.TestingT,
	n, f int,
	peerIdentities []*cryptolib.KeyPair,
	dkShareRegistrys []registry.DKShareRegistry, // Will be used if not nil.
) (*cryptolib.Address, []registry.DKShareRegistry) {
	nodePubKeys := PublicKeys(peerIdentities)
	dssSuite := tcrypto.DefaultEd25519Suite()
	blsSuite := tcrypto.DefaultBLSSuite()
	dssThreshold := n - f
	blsThreshold := f + 1
	dssPubKey, dssPubPoly, dssPriShares := MakeSharedSecret(dssSuite, n, dssThreshold)
	blsPubKey, blsPubPoly, blsPriShares := MakeSharedSecret(blsSuite, n, blsThreshold)
	_, dssCommits := dssPubPoly.Info()
	_, blsCommits := blsPubPoly.Info()
	//
	// Make public shares (do they differ from the commits?)
	dssPublicShares := make([]kyber.Point, n)
	for i := range dssPublicShares {
		dssPublicShares[i] = dssSuite.Point().Mul(dssPriShares[i].V, nil)
	}
	blsPublicShares := make([]kyber.Point, n)
	for i := range blsPublicShares {
		blsPublicShares[i] = blsSuite.Point().Mul(blsPriShares[i].V, nil)
	}
	//
	// Create the DKShare objects.
	if dkShareRegistrys == nil {
		dkShareRegistrys = make([]registry.DKShareRegistry, len(peerIdentities))
	}
	require.Equal(t, n, len(dkShareRegistrys))
	var address *cryptolib.Address
	for i, identity := range peerIdentities {
		indexUint16, err := safecast.Convert[uint16](i)
		require.NoError(t, err)
		nUint16, err := safecast.Convert[uint16](n)
		require.NoError(t, err)
		dssThresholdUint16, err := safecast.Convert[uint16](dssThreshold)
		require.NoError(t, err)
		blsThresholdUint16, err := safecast.Convert[uint16](blsThreshold)
		require.NoError(t, err)

		nodeDKS, err := tcrypto.NewDKShare(
			indexUint16,              // index
			nUint16,                  // n
			dssThresholdUint16,       // t
			identity.GetPrivateKey(), // nodePrivKey
			nodePubKeys,              // nodePubKeys
			dssSuite,                 // edSuite
			dssPubKey,                // edSharedPublic
			dssCommits,               // edPublicCommits
			dssPublicShares,          // edPublicShares
			dssPriShares[i].V,        // edPrivateShare
			blsSuite,                 // blsSuite
			blsThresholdUint16,       // blsThreshold
			blsPubKey,                // blsSharedPublic
			blsCommits,               // blsPublicCommits
			blsPublicShares,          // blsPublicShares
			blsPriShares[i].V,        // blsPrivateShare
		)
		require.NoError(t, err)
		if address == nil {
			address = nodeDKS.GetAddress()
		}
		if dkShareRegistrys[i] == nil {
			dkShareRegistrys[i] = testutil.NewDistributedKeyGenerationRegistry(identity.GetPrivateKey())
		}
		require.NoError(t, dkShareRegistrys[i].SaveDKShare(nodeDKS))
	}
	return address, dkShareRegistrys
}

func MakeSharedSecret(suite suites.Suite, n, t int) (kyber.Point, *share.PubPoly, []*share.PriShare) {
	priPoly := share.NewPriPoly(suite, t, nil, suite.RandomStream())
	priShares := priPoly.Shares(n)
	pubPoly := priPoly.Commit(suite.Point().Base())
	_, commits := pubPoly.Info()
	pubKey := commits[0]
	return pubKey, pubPoly, priShares
}

func SetupNet(
	peeringURLs []string,
	peerIdentities []*cryptolib.KeyPair,
	behavior testutil.PeeringNetBehavior,
	log log.Logger,
) ([]peering.NetworkProvider, io.Closer) {
	peeringNetwork := testutil.NewPeeringNetwork(
		peeringURLs, peerIdentities, 10000, behavior,
		testlogger.WithLevel(log, slog.LevelWarn, false),
	)
	networkProviders := peeringNetwork.NetworkProviders()
	return networkProviders, peeringNetwork
}
