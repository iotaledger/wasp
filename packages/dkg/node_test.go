// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dkg_test

// TODO: Tests with corrupted messages.
// TODO: Tests with byzantine messages.
// TODO: Single node down for some time.

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/dkg"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/util/key"
)

// TestBasic checks if DKG procedure is executed successfully in a common case.
func TestBasic(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()
	//
	// Create a fake network and keys for the tests.
	var timeout = 100 * time.Second
	var threshold uint16 = 10
	var peerCount uint16 = 10
	var peerNetIDs []string = make([]string, peerCount)
	var peerPubs []kyber.Point = make([]kyber.Point, len(peerNetIDs))
	var peerSecs []kyber.Scalar = make([]kyber.Scalar, len(peerNetIDs))
	var suite = pairing.NewSuiteBn256() // NOTE: That's from the Pairing Adapter.
	for i := range peerNetIDs {
		peerPair := key.NewKeyPair(suite)
		peerNetIDs[i] = fmt.Sprintf("P%02d", i)
		peerSecs[i] = peerPair.Private
		peerPubs[i] = peerPair.Public
	}
	var peeringNetwork *testutil.PeeringNetwork = testutil.NewPeeringNetwork(
		peerNetIDs, peerPubs, peerSecs, 10000,
		testutil.NewPeeringNetReliable(),
		testlogger.WithLevel(log, logger.LevelWarn, false),
	)
	var networkProviders []peering.NetworkProvider = peeringNetwork.NetworkProviders()
	//
	// Initialize the DKG subsystem in each node.
	var dkgNodes []*dkg.Node = make([]*dkg.Node, len(peerNetIDs))
	for i := range peerNetIDs {
		registry := testutil.NewDkgRegistryProvider(suite)
		dkgNodes[i] = dkg.NewNode(
			peerSecs[i], peerPubs[i], suite, networkProviders[i], registry,
			testlogger.WithLevel(log.With("NetID", peerNetIDs[i]), logger.LevelDebug, false),
		)
	}
	//
	// Initiate the key generation from some client node.
	dkShare, err := dkgNodes[0].GenerateDistributedKey(
		peerNetIDs,
		peerPubs,
		threshold,
		1*time.Second,
		2*time.Second,
		timeout,
	)
	require.Nil(t, err)
	require.NotNil(t, dkShare.Address)
	require.NotNil(t, dkShare.SharedPublic)
}

// TestNoPubs checks, if public keys are taken from the peering network successfully.
// See a NOTE in the test case bellow.
func TestNoPubs(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()
	//
	// Create a fake network and keys for the tests.
	var timeout = 100 * time.Second
	var threshold uint16 = 10
	var peerCount uint16 = 10
	var peerNetIDs []string = make([]string, peerCount)
	var peerPubs []kyber.Point = make([]kyber.Point, len(peerNetIDs))
	var peerSecs []kyber.Scalar = make([]kyber.Scalar, len(peerNetIDs))
	var suite = pairing.NewSuiteBn256() // That's from the Pairing Adapter.
	for i := range peerNetIDs {
		peerPair := key.NewKeyPair(suite)
		peerNetIDs[i] = fmt.Sprintf("P%02d", i)
		peerSecs[i] = peerPair.Private
		peerPubs[i] = peerPair.Public
	}
	var peeringNetwork *testutil.PeeringNetwork = testutil.NewPeeringNetwork(
		peerNetIDs, peerPubs, peerSecs, 10000,
		testutil.NewPeeringNetReliable(),
		testlogger.WithLevel(log, logger.LevelWarn, false),
	)
	var networkProviders []peering.NetworkProvider = peeringNetwork.NetworkProviders()
	//
	// Initialize the DKG subsystem in each node.
	var dkgNodes []*dkg.Node = make([]*dkg.Node, len(peerNetIDs))
	for i := range peerNetIDs {
		registry := testutil.NewDkgRegistryProvider(suite)
		dkgNodes[i] = dkg.NewNode(
			peerSecs[i], peerPubs[i], suite, networkProviders[i], registry,
			testlogger.WithLevel(log.With("NetID", peerNetIDs[i]), logger.LevelDebug, false),
		)
	}
	//
	// Initiate the key generation from some client node.
	dkShare, err := dkgNodes[0].GenerateDistributedKey(
		peerNetIDs,
		nil, // NOTE: Should be taken from the peering node.
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
	var timeout = 100 * time.Second
	var threshold uint16 = 10
	var peerCount uint16 = 10
	var peerNetIDs []string = make([]string, peerCount)
	var peerPubs []kyber.Point = make([]kyber.Point, len(peerNetIDs))
	var peerSecs []kyber.Scalar = make([]kyber.Scalar, len(peerNetIDs))
	var suite = pairing.NewSuiteBn256() // That's from the Pairing Adapter.
	for i := range peerNetIDs {
		peerPair := key.NewKeyPair(suite)
		peerNetIDs[i] = fmt.Sprintf("P%02d", i)
		peerSecs[i] = peerPair.Private
		peerPubs[i] = peerPair.Public
	}
	var peeringNetwork *testutil.PeeringNetwork = testutil.NewPeeringNetwork(
		peerNetIDs, peerPubs, peerSecs, 10000,
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
		registry := testutil.NewDkgRegistryProvider(suite)
		dkgNodes[i] = dkg.NewNode(
			peerSecs[i], peerPubs[i], suite, networkProviders[i], registry,
			testlogger.WithLevel(log.With("NetID", peerNetIDs[i]), logger.LevelDebug, false),
		)
	}
	//
	// Initiate the key generation from some client node.
	dkShare, err := dkgNodes[0].GenerateDistributedKey(
		peerNetIDs,
		nil, // NOTE: Should be taken from the peering node.
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
		var timeout = 100 * time.Second
		var threshold uint16 = n
		var peerCount uint16 = n
		var peerNetIDs []string = make([]string, peerCount)
		var peerPubs []kyber.Point = make([]kyber.Point, len(peerNetIDs))
		var peerSecs []kyber.Scalar = make([]kyber.Scalar, len(peerNetIDs))
		var suite = pairing.NewSuiteBn256() // NOTE: That's from the Pairing Adapter.
		for i := range peerNetIDs {
			peerPair := key.NewKeyPair(suite)
			peerNetIDs[i] = fmt.Sprintf("P%02d", i)
			peerSecs[i] = peerPair.Private
			peerPubs[i] = peerPair.Public
		}
		var peeringNetwork *testutil.PeeringNetwork = testutil.NewPeeringNetwork(
			peerNetIDs, peerPubs, peerSecs, 10000,
			testutil.NewPeeringNetReliable(),
			testlogger.WithLevel(log, logger.LevelWarn, false),
		)
		var networkProviders []peering.NetworkProvider = peeringNetwork.NetworkProviders()
		//
		// Initialize the DKG subsystem in each node.
		var dkgNodes []*dkg.Node = make([]*dkg.Node, len(peerNetIDs))
		for i := range peerNetIDs {
			registry := testutil.NewDkgRegistryProvider(suite)
			dkgNodes[i] = dkg.NewNode(
				peerSecs[i], peerPubs[i], suite, networkProviders[i], registry,
				testlogger.WithLevel(log.With("NetID", peerNetIDs[i]), logger.LevelDebug, false),
			)
		}
		//
		// Initiate the key generation from some client node.
		dkShare, err := dkgNodes[0].GenerateDistributedKey(
			peerNetIDs,
			peerPubs,
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
