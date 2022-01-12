// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testpeers

import (
	"fmt"
	iotago "github.com/iotaledger/iota.go/v3"
	"io"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/dkg"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/stretchr/testify/require"
)

func SetupKeys(peerCount uint16) ([]string, []*cryptolib.KeyPair) {
	peerNetIDs := make([]string, peerCount)
	peerIdentities := make([]*cryptolib.KeyPair, peerCount)
	for i := range peerNetIDs {
		peerIdentity := cryptolib.NewKeyPair()
		peerNetIDs[i] = fmt.Sprintf("P%02d", i)
		peerIdentities[i] = &peerIdentity
	}
	return peerNetIDs, peerIdentities
}

func PublicKeys(peerIdentities []*cryptolib.KeyPair) []cryptolib.PublicKey {
	pubKeys := make([]cryptolib.PublicKey, len(peerIdentities))
	for i := range pubKeys {
		pubKeys[i] = &peerIdentities[i].PublicKey
	}
	return pubKeys
}

func SetupDkg(
	t *testing.T,
	threshold uint16,
	peerNetIDs []string,
	peerIdentities []*cryptolib.KeyPair,
	suite tcrypto.Suite,
	log *logger.Logger,
) (iotago.Address, []registry.DKShareRegistryProvider) {
	timeout := 100 * time.Second
	networkProviders, networkCloser := SetupNet(peerNetIDs, peerIdentities, testutil.NewPeeringNetReliable(log), log)
	//
	// Initialize the DKG subsystem in each node.
	dkgNodes := make([]*dkg.Node, len(peerNetIDs))
	registries := make([]registry.DKShareRegistryProvider, len(peerNetIDs))
	for i := range peerNetIDs {
		registries[i] = testutil.NewDkgRegistryProvider(suite)
		dkgNode, err := dkg.NewNode(
			peerIdentities[i], networkProviders[i], registries[i],
			testlogger.WithLevel(log.With("NetID", peerNetIDs[i]), logger.LevelError, false),
		)
		require.NoError(t, err)
		dkgNodes[i] = dkgNode
	}
	//
	// Initiate the key generation from some client node.
	dkShare, err := dkgNodes[0].GenerateDistributedKey(
		PublicKeys(peerIdentities),
		threshold,
		100*time.Second,
		200*time.Second,
		timeout,
	)
	require.Nil(t, err)
	require.NotNil(t, dkShare.Address)
	require.NotNil(t, dkShare.SharedPublic)
	require.NoError(t, networkCloser.Close())
	return dkShare.Address, registries
}

func SetupDkgPregenerated(
	t *testing.T,
	threshold uint16,
	identities []*ed25519.KeyPair,
	suite tcrypto.Suite,
) (iotago.Address, []registry.DKShareRegistryProvider) {
	var err error
	var serializedDks [][]byte = pregeneratedDksRead(uint16(len(identities)), threshold)
	nodePubKeys := make([]*ed25519.PublicKey, len(identities))
	for i := range nodePubKeys {
		nodePubKeys[i] = &identities[i].PublicKey
	}
	dks := make([]*tcrypto.DKShare, len(serializedDks))
	registries := make([]registry.DKShareRegistryProvider, len(identities))
	for i := range dks {
		dks[i], err = tcrypto.DKShareFromBytes(serializedDks[i], suite)
		dks[i].NodePubKeys = nodePubKeys
		if i > 0 {
			// It was removed to decrease the serialized size.
			dks[i].PublicCommits = dks[0].PublicCommits
			dks[i].PublicShares = dks[0].PublicShares
		}
		require.Nil(t, err)
		registries[i] = testutil.NewDkgRegistryProvider(suite)
		require.Nil(t, registries[i].SaveDKShare(dks[i]))
	}
	require.Equal(t, dks[0].N, uint16(len(identities)), "dks was pregenerated for different node count (N=%v)", dks[0].N)
	require.Equal(t, dks[0].T, threshold, "dks was pregenerated for different threshold (T=%v)", dks[0].T)
	return dks[0].Address, registries
}

func SetupNet(
	peerNetIDs []string,
	peerIdentities []*cryptolib.KeyPair,
	behavior testutil.PeeringNetBehavior,
	log *logger.Logger,
) ([]peering.NetworkProvider, io.Closer) {
	peeringNetwork := testutil.NewPeeringNetwork(
		peerNetIDs, peerIdentities, 10000, behavior,
		testlogger.WithLevel(log, logger.LevelWarn, false),
	)
	networkProviders := peeringNetwork.NetworkProviders()
	return networkProviders, peeringNetwork
}
