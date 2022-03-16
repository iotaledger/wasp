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
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/nodeconn"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/privtangle"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
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
	chainOICh := make(chan *iotago.UTXOInput)
	chainOuts := make(map[iotago.UTXOInput]*iotago.AliasOutput)
	reqs := make(map[iscp.RequestID]*iscp.OnLedgerRequestData)
	reqIDCh := make(chan iscp.RequestID)
	ncChain := nc.RegisterChain(chainAddr)
	ncChain.AttachToAliasOutput(func(o *iscp.AliasOutputWithID) {
		oi := o.ID()
		chainOuts[*oi] = o.GetAliasOutput()
		chainOICh <- oi
	})
	ncChain.AttachToOnLedgerRequest(func(r *iscp.OnLedgerRequestData) {
		rid := r.ID()
		reqs[rid] = r
		reqIDCh <- rid
	})

	// Post a TX directly, and wait for it in the message stream (e.g. a request).
	_, err := pt.PostSimpleValueTX(ctx, pt.NodeClient(0), pt.FaucetKeyPair, chainAddr, 50000)
	require.NoError(t, err)
	t.Logf("Waiting for outputs posted via tangle...")
	rid := <-reqIDCh
	t.Logf("Waiting for outputs posted via tangle... Done, have %s=%v", rid, reqs[rid].ID())

	// Post a TX via the NodeConn (e.g. alias output).
	tiseCh := make(chan bool)
	ncChain.AttachToTxInclusionState(func(txID iotago.TransactionID, inclusionState string) {
		t.Logf("TX Inclusion state changed, txID=%v, state=%v", iscp.TxID(&txID), inclusionState)
		if inclusionState == "included" {
			tiseCh <- true
		}
	})
	tx, err := pt.MakeSimpleValueTX(ctx, pt.NodeClient(0), chainKeys, chainAddr, 50000)
	require.NoError(t, err)
	err = nc.PublishTransaction(chainAddr, uint32(0), tx)
	require.NoError(t, err)
	t.Logf("Waiting for outputs posted via nodeConn...")
	rid = <-reqIDCh
	t.Logf("Waiting for outputs posted via nodeConn... Done, have %v=%v", rid, reqs[rid].ID())
	t.Logf("Waiting for TX incusion event...")
	<-tiseCh
	t.Logf("Waiting for TX incusion event... Done")

	ncChain.DetachFromAliasOutput()
	ncChain.DetachFromTxInclusionState()
	nc.UnregisterChain(chainAddr)

	//
	// Cleanup.
	require.NoError(t, networkCloser.Close())
}
