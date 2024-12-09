// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainmanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/kvstore/mapdb"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/packages/chain/chainmanager"
	"github.com/iotaledger/wasp/packages/chain/cons"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/origin"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/l1starter"
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
)

func TestChainMgrBasic(t *testing.T) {
	type test struct {
		n int
		f int
	}
	tests := []test{
		{n: 1, f: 0}, // Low N.
		// {n: 2, f: 0},   // Low N. TODO: This is disabled temporarily.
		{n: 3, f: 0},   // Low N.
		{n: 4, f: 1},   // Smallest robust cluster.
		{n: 10, f: 3},  // Typical config.
		{n: 31, f: 10}, // Large cluster.
	}
	for i := range tests {
		tst := tests[i]
		t.Run(
			fmt.Sprintf("N=%v,F=%v", tst.n, tst.f),
			func(tt *testing.T) { testChainMgrBasic(tt, tst.n, tst.f) },
		)
	}
}

func testChainMgrBasic(t *testing.T, n, f int) {
	log := testlogger.NewLogger(t)
	defer log.Sync()
	//
	// Create ledger accounts.
	//utxoDB := utxodb.New(utxodb.DefaultInitParams())
	originator := cryptolib.NewKeyPair()
	//
	// Node identities and DKG.
	_, peerIdentities := testpeers.SetupKeys(uint16(n))
	nodeIDs := make([]gpa.NodeID, len(peerIdentities))
	for i, pid := range peerIdentities {
		nodeIDs[i] = gpa.NodeIDFromPublicKey(pid.GetPublicKey())
	}
	cmtAddrA, dkRegs := testpeers.SetupDkgTrivial(t, n, f, peerIdentities, nil)
	cmtAddrB, dkRegs := testpeers.SetupDkgTrivial(t, n, f, peerIdentities, dkRegs)
	require.NotNil(t, cmtAddrA)
	require.NotNil(t, cmtAddrB)
	//
	// Chain identifiers.
	tcl := newTestChainLedger(t, originator)
	anchor, deposit := tcl.MakeTxChainOrigin(cmtAddrA)
	//
	// Construct the nodes.
	nodes := map[gpa.NodeID]gpa.GPA{}
	stores := map[gpa.NodeID]state.Store{}
	for i, nid := range nodeIDs {
		consensusStateRegistry := testutil.NewConsensusStateRegistry()
		stores[nid] = state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB())
		_, err := origin.InitChainByAnchor(stores[nid], anchor, deposit, isc.BaseTokenCoinInfo)
		require.NoError(t, err)
		activeAccessNodesCB := func() ([]*cryptolib.PublicKey, []*cryptolib.PublicKey) {
			return []*cryptolib.PublicKey{}, []*cryptolib.PublicKey{}
		}
		trackActiveStateCB := func(ao *isc.StateAnchor) {
			// Nothing
		}
		savePreliminaryBlockCB := func(state.Block) {
			// Nothing
		}
		updateCommitteeNodesCB := func(tcrypto.DKShare) {
			// Nothing
		}
		cm, err := chainmanager.New(
			nid, anchor.ChainID(), stores[nid], consensusStateRegistry, dkRegs[i], gpa.NodeIDFromPublicKey,
			activeAccessNodesCB, trackActiveStateCB, savePreliminaryBlockCB, updateCommitteeNodesCB, true, -1, 1, nil,
			log.Named(nid.ShortString()),
		)
		require.NoError(t, err)
		nodes[nid] = cm.AsGPA()
	}
	tc := gpa.NewTestContext(nodes)
	tc.PrintAllStatusStrings("Started", t.Logf)
	//
	// Provide initial AO.
	initAOInputs := map[gpa.NodeID]gpa.Input{}
	for nid := range nodes {
		initAOInputs[nid] = chainmanager.NewInputAliasOutputConfirmed(originator.Address(), anchor)
	}
	tc.WithInputs(initAOInputs)
	tc.RunAll()
	tc.PrintAllStatusStrings("Initial AO received", t.Logf)
	for _, n := range nodes {
		out := n.Output().(*chainmanager.Output)
		require.Equal(t, 0, out.NeedPublishTX().Size())
		require.NotNil(t, out.NeedConsensus())
		require.Equal(t, anchor, out.NeedConsensus().BaseStateAnchor)
		require.Equal(t, uint32(1), out.NeedConsensus().LogIndex.AsUint32())
		require.Equal(t, cmtAddrA, &out.NeedConsensus().CommitteeAddr)
	}
	//
	// Provide consensus output.
	step2AO := tcl.FakeRotationTX(anchor, cmtAddrA)
	step2TX := &iotasigner.SignedTransaction{}
	for nid := range nodes {
		consReq := nodes[nid].Output().(*chainmanager.Output).NeedConsensus()
		fake2ST := indexedstore.NewFake(state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB()))
		origin.InitChain(0, fake2ST, nil, iotago.ObjectID{}, 0, isc.BaseTokenCoinInfo)
		block0, err := fake2ST.BlockByIndex(0)
		require.NoError(t, err)

		// TODO: Commit a block to the store, if needed.
		tc.WithInput(nid, chainmanager.NewInputConsensusOutputDone( // TODO: Consider the SKIP cases as well.
			*cmtAddrA,
			consReq.LogIndex, *consReq.BaseStateAnchor.GetObjectID(),
			&cons.Result{
				Transaction: step2TX,
				Block:       block0,
			},
		))
	}
	tc.RunAll()
	tc.PrintAllStatusStrings("Consensus done", t.Logf)
	for nodeID, n := range nodes {
		out := n.Output().(*chainmanager.Output)
		t.Logf("node=%v should have 1 TX to publish, have out=%v", nodeID, out)
		require.Equal(t, 1, out.NeedPublishTX().Size(), "node=%v should have 1 TX to publish, have out=%v", nodeID, out)
		require.Equal(t, step2TX, func() *iotasigner.SignedTransaction {
			tx, _ := out.NeedPublishTX().Get(step2AO.Hash())
			return tx.Tx
		}())
		require.Equal(t, anchor.GetObjectID(), func() iotago.ObjectID {
			tx, _ := out.NeedPublishTX().Get(step2AO.Hash())
			return *tx.BaseAnchorRef.ObjectID
		}())
		require.Equal(t, cmtAddrA, func() *cryptolib.Address {
			tx, _ := out.NeedPublishTX().Get(step2AO.Hash())
			return &tx.CommitteeAddr
		}())
		require.NotNil(t, out.NeedConsensus())
		require.Equal(t, step2AO, out.NeedConsensus().BaseStateAnchor)
		require.Equal(t, uint32(2), out.NeedConsensus().LogIndex.AsUint32())
		require.Equal(t, cmtAddrA, &out.NeedConsensus().CommitteeAddr)
	}
	//
	// Say TX is published
	for nid := range nodes {
		consReq, _ := nodes[nid].Output().(*chainmanager.Output).NeedPublishTX().Get(step2AO.Hash())
		tc.WithInput(nid, chainmanager.NewInputChainTxPublishResult(consReq.CommitteeAddr, consReq.LogIndex, consReq.BaseAnchorRef.Hash(), nil, true))
	}
	tc.RunAll()
	tc.PrintAllStatusStrings("TX Published", t.Logf)
	for _, n := range nodes {
		out := n.Output().(*chainmanager.Output)
		require.Equal(t, 0, out.NeedPublishTX().Size())
		require.NotNil(t, out.NeedConsensus())
		require.Equal(t, step2AO, out.NeedConsensus().BaseStateAnchor)
		require.Equal(t, uint32(2), out.NeedConsensus().LogIndex.AsUint32())
		require.Equal(t, cmtAddrA, &out.NeedConsensus().CommitteeAddr)
	}
	//
	// Say TX is confirmed.
	for nid := range nodes {
		tc.WithInput(nid, chainmanager.NewInputAliasOutputConfirmed(originator.Address(), step2AO))
	}
	tc.RunAll()
	tc.PrintAllStatusStrings("TX Published and Confirmed", t.Logf)
	for _, n := range nodes {
		out := n.Output().(*chainmanager.Output)
		require.Equal(t, 0, out.NeedPublishTX().Size())
		require.NotNil(t, out.NeedConsensus())
		require.Equal(t, step2AO, out.NeedConsensus().BaseStateAnchor)
		require.Equal(t, uint32(2), out.NeedConsensus().LogIndex.AsUint32())
		require.Equal(t, cmtAddrA, &out.NeedConsensus().CommitteeAddr)
	}
	//
	// Make external committee rotation.
	rotateAO := tcl.FakeRotationTX(step2AO, cmtAddrB)
	for nid := range nodes {
		tc.WithInput(nid, chainmanager.NewInputAliasOutputConfirmed(originator.Address(), rotateAO))
	}
	tc.RunAll()
	tc.PrintAllStatusStrings("After external rotation", t.Logf)
	for _, n := range nodes {
		out := n.Output().(*chainmanager.Output)
		require.Equal(t, 0, out.NeedPublishTX().Size())
		require.NotNil(t, out.NeedConsensus())
		require.Equal(t, rotateAO, out.NeedConsensus().BaseStateAnchor)
		require.Equal(t, uint32(1), out.NeedConsensus().LogIndex.AsUint32())
		require.Equal(t, cmtAddrB, &out.NeedConsensus().CommitteeAddr)
	}
}

func newTestChainLedger(t *testing.T, originator *cryptolib.KeyPair) *testchain.TestChainLedger {
	l1client := l1starter.Instance().L1Client()
	l1client.RequestFunds(context.Background(), *originator.Address())
	l1client.RequestFunds(context.Background(), *originator.Address())

	iscPackage, err := l1client.DeployISCContracts(context.Background(), cryptolib.SignerToIotaSigner(originator))
	require.NoError(t, err)

	return testchain.NewTestChainLedger(t, originator, &iscPackage, l1client)
}
