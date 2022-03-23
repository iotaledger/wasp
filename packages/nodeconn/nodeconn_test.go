// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package nodeconn_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	iotagox "github.com/iotaledger/iota.go/v3/x"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/nodeconn"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/privtangle"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/require"
)

func TestNodeConn(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()
	peerCount := 1
	ctx := context.Background()

	//
	// Start the private L1 tangle.
	pt := privtangle.Start(ctx, filepath.Join(os.TempDir(), "wasp-hornet-TestNodeConn"), 16400, 2, t)
	time.Sleep(3 * time.Second)
	t.Logf("Private tangle created.")

	//
	// Start a peering network.
	// peeringID := peering.RandomPeeringID()
	peerNetIDs, peerIdentities := testpeers.SetupKeys(uint16(peerCount))
	networkLog := testlogger.WithLevel(log.Named("Network"), logger.LevelInfo, false)
	networkProviders, networkCloser := testpeers.SetupNet(
		peerNetIDs,
		peerIdentities,
		testutil.NewPeeringNetReliable(networkLog),
		networkLog,
	)
	t.Logf("Peering network created.")

	nc := nodeconn.New("localhost", pt.NodePortRestAPI(0), networkProviders[0], log)
	t.Cleanup(nc.Close)

	//
	// Check milestone attach/detach.
	mChan := make(chan *iotagox.MilestonePointer, 10)
	mSub := nc.AttachMilestones(func(m *iotagox.MilestonePointer) {
		mChan <- m
	})
	<-mChan
	nc.DetachMilestones(mSub)

	//
	// Check the chain operations.
	chainKeys := cryptolib.NewKeyPair()
	chainAddr := chainKeys.GetPublicKey().AsEd25519Address()
	chainOICh := make(chan iotago.OutputID)
	chainOuts := make(map[iotago.OutputID]iotago.Output)
	nc.RegisterChain(chainAddr, func(oi iotago.OutputID, o iotago.Output) {
		chainOuts[oi] = o
		chainOICh <- oi
	})

	client := cluster.NewL1Client(pt.L1Config())
	// Post a TX directly, and wait for it in the message stream (e.g. a request).
	_, err := client.(*cluster.L1Client).PostSimpleValueTX(pt.FaucetKeyPair, chainAddr, 50000)
	require.NoError(t, err)
	t.Logf("Waiting for outputs posted via tangle...")
	oid := <-chainOICh
	t.Logf("Waiting for outputs posted via tangle... Done, have %v=%v", oid.ToHex(), chainOuts[oid])

	// Post a TX via the NodeConn (e.g. alias output).
	tiseCh := make(chan bool)
	tise, err := nc.AttachTxInclusionStateEvents(chainAddr, func(txID iotago.TransactionID, inclusionState string) {
		t.Logf("TX Inclusion state changed, txID=%v, state=%v", txID, inclusionState)
		if inclusionState == "included" {
			tiseCh <- true
		}
	})
	require.NoError(t, err)
	tx, err := client.(*cluster.L1Client).MakeSimpleValueTX(chainKeys, chainAddr, 50000)
	require.NoError(t, err)
	err = nc.PublishTransaction(chainAddr, uint32(0), tx)
	require.NoError(t, err)
	t.Logf("Waiting for outputs posted via nodeConn...")
	oid = <-chainOICh
	t.Logf("Waiting for outputs posted via nodeConn... Done, have %v=%v", oid.ToHex(), chainOuts[oid])
	t.Logf("Waiting for TX incusion event...")
	<-tiseCh
	t.Logf("Waiting for TX incusion event... Done")

	nc.DetachTxInclusionStateEvents(chainAddr, tise)
	nc.UnregisterChain(chainAddr)

	//
	// Cleanup.
	require.NoError(t, networkCloser.Close())
}
