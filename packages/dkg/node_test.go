// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dkg_test

// TODO: Tests with corrupted messages.
// TODO: Tests with byzantine messages.
// TODO: Single node down for some time.

import (
	"testing"
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/dkg"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"github.com/stretchr/testify/require"
)

// TestBasic checks if DKG procedure is executed successfully in a common case.
func TestBasic(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()
	//
	// Create a fake network and keys for the tests.
	timeout := 100 * time.Second
	var threshold uint16 = 10
	var peerCount uint16 = 10
	peerNetIDs, peerIdentities := testpeers.SetupKeys(peerCount)
	var peeringNetwork *testutil.PeeringNetwork = testutil.NewPeeringNetwork(
		peerNetIDs, peerIdentities, 10000,
		testutil.NewPeeringNetReliable(log),
		testlogger.WithLevel(log, logger.LevelWarn, false),
	)
	var networkProviders []peering.NetworkProvider = peeringNetwork.NetworkProviders()
	//
	// Initialize the DKG subsystem in each node.
	var dkgNodes []*dkg.Node = make([]*dkg.Node, len(peerNetIDs))
	for i := range peerNetIDs {
		registry := testutil.NewDkgRegistryProvider(tcrypto.DefaultSuite())
		dkgNode, err := dkg.NewNode(
			peerIdentities[i], networkProviders[i], registry,
			testlogger.WithLevel(log.With("NetID", peerNetIDs[i]), logger.LevelDebug, false),
		)
		require.NoError(t, err)
		dkgNodes[i] = dkgNode
	}
	//
	// Initiate the key generation from some client node.
	dkShare, err := dkgNodes[0].GenerateDistributedKey(
		testpeers.PublicKeys(peerIdentities),
		threshold,
		1*time.Second,
		2*time.Second,
		timeout,
	)
	require.Nil(t, err)
	require.NotNil(t, dkShare.Address)
	require.NotNil(t, dkShare.SharedPublic)
}

// TestUnreliableNet checks, if DKG runs on an unreliable network.
// See a NOTE in the test case bellow.
func TestUnreliableNet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	log := testlogger.NewLogger(t)
	defer log.Sync()
	//
	// Create a fake network and keys for the tests.
	timeout := 100 * time.Second
	var threshold uint16 = 10
	var peerCount uint16 = 10
	peerNetIDs, peerIdentities := testpeers.SetupKeys(peerCount)
	var peeringNetwork *testutil.PeeringNetwork = testutil.NewPeeringNetwork(
		peerNetIDs, peerIdentities, 10000,
		testutil.NewPeeringNetUnreliable( // NOTE: Network parameters.
			80,                                         // Delivered %
			20,                                         // Duplicated %
			10*time.Millisecond, 1000*time.Millisecond, // Delays (from, till)
			testlogger.WithLevel(log.Named("UnreliableNet"), logger.LevelDebug, false),
		),
		testlogger.WithLevel(log, logger.LevelInfo, false),
	)
	var networkProviders []peering.NetworkProvider = peeringNetwork.NetworkProviders()
	//
	// Initialize the DKG subsystem in each node.
	var dkgNodes []*dkg.Node = make([]*dkg.Node, len(peerNetIDs))
	for i := range peerNetIDs {
		registry := testutil.NewDkgRegistryProvider(tcrypto.DefaultSuite())
		dkgNode, err := dkg.NewNode(
			peerIdentities[i], networkProviders[i], registry,
			testlogger.WithLevel(log.With("NetID", peerNetIDs[i]), logger.LevelDebug, false),
		)
		require.NoError(t, err)
		dkgNodes[i] = dkgNode
	}
	//
	// Initiate the key generation from some client node.
	dkShare, err := dkgNodes[0].GenerateDistributedKey(
		testpeers.PublicKeys(peerIdentities),
		threshold,
		100*time.Millisecond, // Round retry.
		500*time.Millisecond, // Step retry.
		timeout,
	)
	require.Nil(t, err)
	require.NotNil(t, dkShare.Address)
	require.NotNil(t, dkShare.SharedPublic)
}

// TestLowN checks, if the DKG works with N=1 and other low values. N=1 is a special case.
func TestLowN(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()
	//
	// Create a fake network and keys for the tests.
	for n := uint16(1); n < 4; n++ {
		timeout := 100 * time.Second
		threshold := n
		peerCount := n
		peerNetIDs, peerIdentities := testpeers.SetupKeys(peerCount)
		var peeringNetwork *testutil.PeeringNetwork = testutil.NewPeeringNetwork(
			peerNetIDs, peerIdentities, 10000,
			testutil.NewPeeringNetReliable(log),
			testlogger.WithLevel(log, logger.LevelWarn, false),
		)
		var networkProviders []peering.NetworkProvider = peeringNetwork.NetworkProviders()
		//
		// Initialize the DKG subsystem in each node.
		var dkgNodes []*dkg.Node = make([]*dkg.Node, len(peerNetIDs))
		for i := range peerNetIDs {
			registry := testutil.NewDkgRegistryProvider(tcrypto.DefaultSuite())
			dkgNode, err := dkg.NewNode(
				peerIdentities[i], networkProviders[i], registry,
				testlogger.WithLevel(log.With("NetID", peerNetIDs[i]), logger.LevelDebug, false),
			)
			require.NoError(t, err)
			dkgNodes[i] = dkgNode
		}
		//
		// Initiate the key generation from some client node.
		dkShare, err := dkgNodes[0].GenerateDistributedKey(
			testpeers.PublicKeys(peerIdentities),
			threshold,
			1*time.Second,
			2*time.Second,
			timeout,
		)
		require.Nil(t, err)
		require.NotNil(t, dkShare.Address)
		require.NotNil(t, dkShare.SharedPublic)
	}
}
