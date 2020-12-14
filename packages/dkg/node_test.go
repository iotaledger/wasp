// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dkg_test

// TODO: Test with unreliable network.

import (
	"fmt"
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

// TestBn256 checks if DKG procedure is executed successfully in a common case.
func TestBn256(t *testing.T) {
	log := testutil.NewLogger(t)
	defer log.Sync()
	t.SkipNow()
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
		peerNetIDs[i] = fmt.Sprintf("P%06d", i)
		peerSecs[i] = peerPair.Private
		peerPubs[i] = peerPair.Public
	}
	var peeringNetwork *testutil.PeeringNetwork = testutil.NewPeeringNetwork(
		peerNetIDs, peerPubs, peerSecs, 10000,
		testutil.WithLevel(log, logger.LevelWarn, true),
	)
	var networkProviders []peering.NetworkProvider = peeringNetwork.NetworkProviders()
	//
	// Initialize the DKG subsystem in each node.
	var dkgNodes []*dkg.Node = make([]*dkg.Node, len(peerNetIDs))
	for i := range peerNetIDs {
		registry := testutil.NewDkgRegistryProvider(suite)
		dkgNodes[i] = dkg.NewNode(
			peerSecs[i], peerPubs[i], suite, networkProviders[i], registry,
			log.With("NetID", peerNetIDs[i]),
		)
	}
	//
	// Initiate the key generation from some client node.
	dkShare, err := dkgNodes[0].GenerateDistributedKey(
		peerNetIDs,
		peerPubs,
		threshold,
		timeout,
	)
	require.Nil(t, err)
	require.NotNil(t, dkShare.ChainID)
	require.NotNil(t, dkShare.Address)
	require.NotNil(t, dkShare.SharedPublic)
}

// TestBn256NoPubs checks, if public keys are taken from the peering network successfully.
// See a NOTE in the test case bellow.
func TestBn256NoPubs(t *testing.T) {
	log := testutil.NewLogger(t)
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
		peerNetIDs[i] = fmt.Sprintf("P%06d", i)
		peerSecs[i] = peerPair.Private
		peerPubs[i] = peerPair.Public
	}
	var peeringNetwork *testutil.PeeringNetwork = testutil.NewPeeringNetwork(
		peerNetIDs, peerPubs, peerSecs, 10000,
		testutil.WithLevel(log, logger.LevelWarn, true),
	)
	var networkProviders []peering.NetworkProvider = peeringNetwork.NetworkProviders()
	//
	// Initialize the DKG subsystem in each node.
	var dkgNodes []*dkg.Node = make([]*dkg.Node, len(peerNetIDs))
	for i := range peerNetIDs {
		registry := testutil.NewDkgRegistryProvider(suite)
		dkgNodes[i] = dkg.NewNode(
			peerSecs[i], peerPubs[i], suite, networkProviders[i], registry,
			log.With("NetID", peerNetIDs[i]),
		)
	}
	//
	// Initiate the key generation from some client node.
	dkShare, err := dkgNodes[0].GenerateDistributedKey(
		peerNetIDs,
		nil, // NOTE: Should be taken from the peering node.
		threshold,
		timeout,
	)
	require.Nil(t, err)
	require.NotNil(t, dkShare.ChainID)
	require.NotNil(t, dkShare.Address)
	require.NotNil(t, dkShare.SharedPublic)
}
