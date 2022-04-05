// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testpeers

import (
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
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
		peerIdentities[i] = cryptolib.NewKeyPair()
		peerNetIDs[i] = fmt.Sprintf("P%02d", i)
	}
	return peerNetIDs, peerIdentities
}

func PublicKeys(peerIdentities []*cryptolib.KeyPair) []*cryptolib.PublicKey {
	pubKeys := make([]*cryptolib.PublicKey, len(peerIdentities))
	for i := range pubKeys {
		pubKeys[i] = peerIdentities[i].GetPublicKey()
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
		registries[i] = testutil.NewDkgRegistryProvider()
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
	require.NotNil(t, dkShare.GetAddress())
	require.NotNil(t, dkShare.GetSharedPublic())
	require.NoError(t, networkCloser.Close())
	return dkShare.GetAddress(), registries
}

func SetupDkgPregenerated(
	t *testing.T,
	threshold uint16,
	identities []*cryptolib.KeyPair,
	suite tcrypto.Suite,
) (iotago.Address, []registry.DKShareRegistryProvider) {
	var err error
	serializedDks := pregeneratedDksRead(uint16(len(identities)), threshold)
	nodePubKeys := make([]*cryptolib.PublicKey, len(identities))
	for i := range nodePubKeys {
		nodePubKeys[i] = identities[i].GetPublicKey()
	}
	dks := make([]*tcrypto.DKShareImpl, len(serializedDks))
	dks[0].AssignNodePubKeys(nodePubKeys)
	registries := make([]registry.DKShareRegistryProvider, len(identities))
	for i := range dks {
		dks[i], err = tcrypto.DKShareFromBytes(serializedDks[i], tcrypto.DefaultEd25519Suite(), tcrypto.DefaultBlsSuite())
		if i > 0 {
			dks[i].AssignCommonData(dks[0])
		}
		require.Nil(t, err)
		registries[i] = testutil.NewDkgRegistryProvider()
		require.Nil(t, registries[i].SaveDKShare(dks[i]))
	}
	require.Equal(t, dks[0].GetN(), uint16(len(identities)), "dks was pregenerated for different node count (N=%v)", dks[0].GetN())
	require.Equal(t, dks[0].GetT(), threshold, "dks was pregenerated for different threshold (T=%v)", dks[0].GetT())
	return dks[0].GetAddress(), registries
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
