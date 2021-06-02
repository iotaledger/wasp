// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testpeers

import (
	"fmt"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/dkg"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/mr-tron/base58"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/util/key"
)

func SetupKeys(peerCount uint16, suite *pairing.SuiteBn256) ([]string, []kyber.Point, []kyber.Scalar) {
	peerNetIDs := make([]string, peerCount)
	peerPubs := make([]kyber.Point, len(peerNetIDs))
	peerSecs := make([]kyber.Scalar, len(peerNetIDs))
	for i := range peerNetIDs {
		peerPair := key.NewKeyPair(suite)
		peerNetIDs[i] = fmt.Sprintf("P%02d", i)
		peerSecs[i] = peerPair.Private
		peerPubs[i] = peerPair.Public
	}
	return peerNetIDs, peerPubs, peerSecs
}

func SetupDkg(
	t *testing.T,
	threshold uint16,
	peerNetIDs []string,
	peerPubs []kyber.Point,
	peerSecs []kyber.Scalar,
	suite *pairing.SuiteBn256,
	log *logger.Logger,
) (ledgerstate.Address, []coretypes.DKShareRegistryProvider) {
	timeout := 1000 * time.Second
	networkProviders := SetupNet(peerNetIDs, peerPubs, peerSecs, testutil.NewPeeringNetReliable(), log)
	//
	// Initialize the DKG subsystem in each node.
	dkgNodes := make([]*dkg.Node, len(peerNetIDs))
	registries := make([]coretypes.DKShareRegistryProvider, len(peerNetIDs))
	for i := range peerNetIDs {
		registries[i] = testutil.NewDkgRegistryProvider(suite)
		dkgNodes[i] = dkg.NewNode(
			peerSecs[i], peerPubs[i], suite, networkProviders[i], registries[i],
			testlogger.WithLevel(log.With("NetID", peerNetIDs[i]), logger.LevelError, false),
		)
	}
	//
	// Initiate the key generation from some client node.
	dkShare, err := dkgNodes[0].GenerateDistributedKey(
		peerNetIDs,
		peerPubs,
		threshold,
		100*time.Second,
		200*time.Second,
		timeout,
	)
	require.Nil(t, err)
	require.NotNil(t, dkShare.Address)
	require.NotNil(t, dkShare.SharedPublic)
	return dkShare.Address, registries
}

func SetupDkgPregenerated(
	t *testing.T,
	threshold uint16,
	peerNetIDs []string,
	suite *pairing.SuiteBn256,
) (ledgerstate.Address, []coretypes.DKShareRegistryProvider) {
	var err error
	var serializedDks []string
	switch len(peerNetIDs) {
	case 1:
		serializedDks = pregeneratedDks1()
	case 4:
		serializedDks = pregeneratedDks4()
	case 10:
		serializedDks = pregeneratedDks10()
	case 40:
		serializedDks = pregeneratedDks40()
	case 70:
		serializedDks = pregeneratedDks70()
	case 100:
		serializedDks = pregeneratedDks100()
	default:
		t.Fatalf("have no keys pregenerated for N=%v", len(peerNetIDs))
	}
	dks := make([]*tcrypto.DKShare, len(serializedDks))
	registries := make([]coretypes.DKShareRegistryProvider, len(peerNetIDs))
	for i := range dks {
		var dksBytes []byte
		dksBytes, err = base58.Decode(serializedDks[i])
		require.Nil(t, err)
		dks[i], err = tcrypto.DKShareFromBytes(dksBytes, suite)
		require.Nil(t, err)
		registries[i] = testutil.NewDkgRegistryProvider(suite)
		require.Nil(t, registries[i].SaveDKShare(dks[i]))
	}
	require.Equal(t, dks[0].T, threshold, "dks was pregenerated for different threshold (T=%v)", dks[0].T)
	return dks[0].Address, registries
}

func SetupNet(
	peerNetIDs []string,
	peerPubs []kyber.Point,
	peerSecs []kyber.Scalar,
	behavior testutil.PeeringNetBehavior,
	log *logger.Logger,
) []peering.NetworkProvider {
	peeringNetwork := testutil.NewPeeringNetwork(
		peerNetIDs, peerPubs, peerSecs, 10000, behavior,
		testlogger.WithLevel(log, logger.LevelWarn, false),
	)
	return peeringNetwork.NetworkProviders()
}
