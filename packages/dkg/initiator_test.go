package dkg_test

// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import (
	"fmt"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/dkg"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/group/edwards25519"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/util/key"
)

func TestEd25519(t *testing.T) {
	log := testutil.NewLogger(t)
	defer log.Sync()
	//
	// Create a fake network and keys for the tests.
	var timeout = 100 * time.Second
	var threshold uint16 = 10
	var peerCount = 10
	var peerLocs []string = make([]string, peerCount)
	var peerPubs []kyber.Point = make([]kyber.Point, len(peerLocs))
	var peerSecs []kyber.Scalar = make([]kyber.Scalar, len(peerLocs))
	var suite = edwards25519.NewBlakeSHA256Ed25519()
	for i := range peerLocs {
		peerLocs[i] = fmt.Sprintf("P%06d", i)
		peerSecs[i] = suite.Scalar().Pick(suite.RandomStream())
		peerPubs[i] = suite.Point().Mul(peerSecs[i], nil)
	}
	var peeringNetwork *testutil.PeeringNetwork = testutil.NewPeeringNetwork(
		peerLocs, peerPubs, peerSecs, 10000,
		testutil.WithLevel(log, logger.LevelWarn),
	)
	var networkProviders []peering.NetworkProvider = peeringNetwork.NetworkProviders()
	//
	// Initialize the DKG subsystem in each node.
	var dkgNodes []*dkg.Node = make([]*dkg.Node, len(peerLocs))
	for i := range peerLocs {
		registry := testutil.NewDkgRegistryProvider(suite)
		dkgNodes[i] = dkg.Init(
			peerSecs[i], peerPubs[i], suite, networkProviders[i], registry,
			log.With("loc", peerLocs[i]),
		)
	}
	//
	// Initiate the key generation from some client node.
	var initiatorKey = suite.Scalar().Pick(suite.RandomStream())
	var initiatorPub = suite.Point().Mul(initiatorKey, nil)
	sharedAddr, sharedPub, err := dkg.GenerateDistributedKey(&dkg.GenerateDistributedKeyParams{
		InitiatorPub: initiatorPub,
		PeerLocs:     peerLocs,
		PeerPubs:     peerPubs,
		Threshold:    threshold,
		Version:      address.VersionED25519,
		Timeout:      timeout,
		Suite:        suite,
		NetProvider:  networkProviders[0],
		Logger:       log.Named("initiator"),
	})
	require.Nil(t, err)
	require.NotNil(t, sharedAddr)
	require.NotNil(t, sharedPub)
}

func TestBn256(t *testing.T) {
	log := testutil.NewLogger(t)
	defer log.Sync()
	//
	// Create a fake network and keys for the tests.
	var timeout = 100 * time.Second
	var threshold uint16 = 10
	var peerCount uint16 = 10
	var peerLocs []string = make([]string, peerCount)
	var peerPubs []kyber.Point = make([]kyber.Point, len(peerLocs))
	var peerSecs []kyber.Scalar = make([]kyber.Scalar, len(peerLocs))
	var suite = pairing.NewSuiteBn256() // NOTE: That's from the Pairing Adapter.
	for i := range peerLocs {
		peerPair := key.NewKeyPair(suite)
		peerLocs[i] = fmt.Sprintf("P%06d", i)
		peerSecs[i] = peerPair.Private
		peerPubs[i] = peerPair.Public
	}
	var peeringNetwork *testutil.PeeringNetwork = testutil.NewPeeringNetwork(
		peerLocs, peerPubs, peerSecs, 10000,
		testutil.WithLevel(log, logger.LevelWarn),
	)
	var networkProviders []peering.NetworkProvider = peeringNetwork.NetworkProviders()
	//
	// Initialize the DKG subsystem in each node.
	var dkgNodes []*dkg.Node = make([]*dkg.Node, len(peerLocs))
	for i := range peerLocs {
		registry := testutil.NewDkgRegistryProvider(suite)
		dkgNodes[i] = dkg.Init(
			peerSecs[i], peerPubs[i], suite, networkProviders[i], registry,
			log.With("loc", peerLocs[i]),
		)
	}
	//
	// Initiate the key generation from some client node.
	var initiatorPair = key.NewKeyPair(suite)
	sharedAddr, sharedPub, err := dkg.GenerateDistributedKey(&dkg.GenerateDistributedKeyParams{
		InitiatorPub: initiatorPair.Public,
		PeerLocs:     peerLocs,
		PeerPubs:     peerPubs,
		Threshold:    threshold,
		Version:      address.VersionED25519,
		Timeout:      timeout,
		Suite:        suite,
		NetProvider:  networkProviders[0],
		Logger:       log.Named("initiator"),
	})
	require.Nil(t, err)
	require.NotNil(t, sharedAddr)
	require.NotNil(t, sharedPub)
}
