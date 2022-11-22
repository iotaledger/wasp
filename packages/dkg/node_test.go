// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dkg_test

// TODO: Tests with corrupted messages.
// TODO: Tests with byzantine messages.
// TODO: Single node down for some time.

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/dkg"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
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
	peeringNetwork := testutil.NewPeeringNetwork(
		peerNetIDs, peerIdentities, 10000,
		testutil.NewPeeringNetReliable(log),
		testlogger.WithLevel(log, logger.LevelWarn, false),
	)
	networkProviders := peeringNetwork.NetworkProviders()
	//
	// Initialize the DKG subsystem in each node.
	dkgNodes := make([]*dkg.Node, len(peerNetIDs))
	dkShareRegistryProviders := make([]registry.DKShareRegistryProvider, len(peerNetIDs))
	for i := range peerNetIDs {
		dkShareRegistryProviders[i] = testutil.NewDkgRegistryProvider(peerIdentities[i].GetPrivateKey())
		dkgNode, err := dkg.NewNode(
			peerIdentities[i], networkProviders[i], dkShareRegistryProviders[i],
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
	require.NotNil(t, dkShare.GetAddress())
	require.NotNil(t, dkShare.GetSharedPublic())
	//
	// Aggregate the signatures: generate signature shares.
	dataToSign := []byte{112, 117, 116, 105, 110, 32, 99, 104, 117, 105, 108, 111, 33}
	require.NoError(t, err)
	// dssPartSigs := make([]*dss.PartialSig, len(peerNetIDs))
	blsPartSigs := make([][]byte, len(peerNetIDs))
	var aggrDks tcrypto.DKShare
	for i, r := range dkShareRegistryProviders {
		dks, err := r.LoadDKShare(dkShare.GetAddress())
		if i == 0 {
			aggrDks = dks
		}
		require.NoError(t, err)
		// dssPartSigs[i], err = dks.DSSSignShare(dataToSign) // TODO: Check the signature.
		// require.NoError(t, err)
		blsPartSigs[i], err = dks.BLSSignShare(dataToSign)
		require.NoError(t, err)
	}
	//
	// Aggregate the signatures: check the DSS signature. // TODO: Check the signature.
	// dssAggrSig, err := aggrDks.DSSRecoverMasterSignature(dssPartSigs, dataToSign)
	// require.NoError(t, err)
	// require.NotNil(t, dssAggrSig)
	// require.True(t, aggrDks.GetSharedPublic().Verify(dataToSign, dssAggrSig))
	//
	// Aggregate the signatures: check the BLS signature.
	blsAggrSig, err := aggrDks.BLSRecoverMasterSignature(blsPartSigs, dataToSign)
	require.NoError(t, err)
	require.NotNil(t, blsAggrSig)
	require.NoError(t, aggrDks.BLSVerifyMasterSignature(dataToSign, blsAggrSig.Signature[:]))
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
	peeringNetwork := testutil.NewPeeringNetwork(
		peerNetIDs, peerIdentities, 10000,
		testutil.NewPeeringNetUnreliable( // NOTE: Network parameters.
			80,                                         // Delivered %
			20,                                         // Duplicated %
			10*time.Millisecond, 1000*time.Millisecond, // Delays (from, till)
			testlogger.WithLevel(log.Named("UnreliableNet"), logger.LevelDebug, false),
		),
		testlogger.WithLevel(log, logger.LevelInfo, false),
	)
	networkProviders := peeringNetwork.NetworkProviders()
	//
	// Initialize the DKG subsystem in each node.
	dkgNodes := make([]*dkg.Node, len(peerNetIDs))
	for i := range peerNetIDs {
		dksReg := testutil.NewDkgRegistryProvider(peerIdentities[i].GetPrivateKey())
		dkgNode, err := dkg.NewNode(
			peerIdentities[i], networkProviders[i], dksReg,
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
	require.NotNil(t, dkShare.GetAddress())
	require.NotNil(t, dkShare.GetSharedPublic())
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
		peeringNetwork := testutil.NewPeeringNetwork(
			peerNetIDs, peerIdentities, 10000,
			testutil.NewPeeringNetReliable(log),
			testlogger.WithLevel(log, logger.LevelWarn, false),
		)
		networkProviders := peeringNetwork.NetworkProviders()
		//
		// Initialize the DKG subsystem in each node.
		dkgNodes := make([]*dkg.Node, len(peerNetIDs))
		for i := range peerNetIDs {
			dksReg := testutil.NewDkgRegistryProvider(peerIdentities[i].GetPrivateKey())
			dkgNode, err := dkg.NewNode(
				peerIdentities[i], networkProviders[i], dksReg,
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
		require.NotNil(t, dkShare.GetAddress())
		require.NotNil(t, dkShare.GetSharedPublic())
	}
}
