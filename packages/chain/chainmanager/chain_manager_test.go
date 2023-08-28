// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainmanager_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	iotago "github.com/iotaledger/iota.go/v3"
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
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"github.com/iotaledger/wasp/packages/testutil/utxodb"
)

func TestChainMgrBasic(t *testing.T) {
	t.Skip("flaky")
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
	utxoDB := utxodb.New(utxodb.DefaultInitParams())
	originator := cryptolib.NewKeyPair()
	_, err := utxoDB.GetFundsFromFaucet(originator.Address())
	require.NoError(t, err)
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
	tcl := testchain.NewTestChainLedger(t, utxoDB, originator)
	_, originAO, chainID := tcl.MakeTxChainOrigin(cmtAddrA)
	//
	// Construct the nodes.
	nodes := map[gpa.NodeID]gpa.GPA{}
	stores := map[gpa.NodeID]state.Store{}
	for i, nid := range nodeIDs {
		consensusStateRegistry := testutil.NewConsensusStateRegistry()
		stores[nid] = state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB())
		_, err := origin.InitChainByAliasOutput(stores[nid], originAO)
		require.NoError(t, err)
		activeAccessNodesCB := func() ([]*cryptolib.PublicKey, []*cryptolib.PublicKey) {
			return []*cryptolib.PublicKey{}, []*cryptolib.PublicKey{}
		}
		trackActiveStateCB := func(ao *isc.AliasOutputWithID) {
			// Nothing
		}
		savePreliminaryBlockCB := func(state.Block) {
			// Nothing
		}
		updateCommitteeNodesCB := func(tcrypto.DKShare) {
			// Nothing
		}
		cm, err := chainmanager.New(
			nid, chainID, stores[nid], consensusStateRegistry, dkRegs[i], gpa.NodeIDFromPublicKey,
			activeAccessNodesCB, trackActiveStateCB, savePreliminaryBlockCB, updateCommitteeNodesCB, true, -1, nil,
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
		initAOInputs[nid] = chainmanager.NewInputAliasOutputConfirmed(originAO)
	}
	tc.WithInputs(initAOInputs)
	tc.RunAll()
	tc.PrintAllStatusStrings("Initial AO received", t.Logf)
	for _, n := range nodes {
		out := n.Output().(*chainmanager.Output)
		require.Equal(t, 0, out.NeedPublishTX().Size())
		require.NotNil(t, out.NeedConsensus())
		require.Equal(t, originAO, out.NeedConsensus().BaseAliasOutput)
		require.Equal(t, uint32(1), out.NeedConsensus().LogIndex.AsUint32())
		require.Equal(t, cmtAddrA, &out.NeedConsensus().CommitteeAddr)
	}
	//
	// Provide consensus output.
	step2AO, step2TX := tcl.FakeRotationTX(originAO, cmtAddrA)
	for nid := range nodes {
		consReq := nodes[nid].Output().(*chainmanager.Output).NeedConsensus()
		fake2ST := indexedstore.NewFake(state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB()))
		origin.InitChain(fake2ST, nil, 0)
		block0, err := fake2ST.BlockByIndex(0)
		require.NoError(t, err)
		// TODO: Commit a block to the store, if needed.
		tc.WithInput(nid, chainmanager.NewInputConsensusOutputDone( // TODO: Consider the SKIP cases as well.
			*cmtAddrA.(*iotago.Ed25519Address),
			consReq.LogIndex, consReq.BaseAliasOutput.OutputID(),
			&cons.Result{
				Transaction:     step2TX,
				Block:           block0,
				BaseAliasOutput: consReq.BaseAliasOutput.OutputID(),
				NextAliasOutput: step2AO,
			},
		))
	}
	tc.RunAll()
	tc.PrintAllStatusStrings("Consensus done", t.Logf)
	for nodeID, n := range nodes {
		out := n.Output().(*chainmanager.Output)
		t.Logf("node=%v should have 1 TX to publish, have out=%v", nodeID, out)
		require.Equal(t, 1, out.NeedPublishTX().Size(), "node=%v should have 1 TX to publish, have out=%v", nodeID, out)
		require.Equal(t, step2TX, func() *iotago.Transaction { tx, _ := out.NeedPublishTX().Get(step2AO.TransactionID()); return tx.Tx }())
		require.Equal(t, originAO.OutputID(), func() iotago.OutputID {
			tx, _ := out.NeedPublishTX().Get(step2AO.TransactionID())
			return tx.BaseAliasOutputID
		}())
		require.Equal(t, cmtAddrA, func() iotago.Address {
			tx, _ := out.NeedPublishTX().Get(step2AO.TransactionID())
			return &tx.CommitteeAddr
		}())
		require.NotNil(t, out.NeedConsensus())
		require.Equal(t, step2AO, out.NeedConsensus().BaseAliasOutput)
		require.Equal(t, uint32(2), out.NeedConsensus().LogIndex.AsUint32())
		require.Equal(t, cmtAddrA, &out.NeedConsensus().CommitteeAddr)
	}
	//
	// Say TX is published
	for nid := range nodes {
		consReq, _ := nodes[nid].Output().(*chainmanager.Output).NeedPublishTX().Get(step2AO.TransactionID())
		tc.WithInput(nid, chainmanager.NewInputChainTxPublishResult(consReq.CommitteeAddr, consReq.LogIndex, consReq.TxID, consReq.NextAliasOutput, true))
	}
	tc.RunAll()
	tc.PrintAllStatusStrings("TX Published", t.Logf)
	for _, n := range nodes {
		out := n.Output().(*chainmanager.Output)
		require.Equal(t, 0, out.NeedPublishTX().Size())
		require.NotNil(t, out.NeedConsensus())
		require.Equal(t, step2AO, out.NeedConsensus().BaseAliasOutput)
		require.Equal(t, uint32(2), out.NeedConsensus().LogIndex.AsUint32())
		require.Equal(t, cmtAddrA, &out.NeedConsensus().CommitteeAddr)
	}
	//
	// Say TX is confirmed.
	for nid := range nodes {
		tc.WithInput(nid, chainmanager.NewInputAliasOutputConfirmed(step2AO))
	}
	tc.RunAll()
	tc.PrintAllStatusStrings("TX Published and Confirmed", t.Logf)
	for _, n := range nodes {
		out := n.Output().(*chainmanager.Output)
		require.Equal(t, 0, out.NeedPublishTX().Size())
		require.NotNil(t, out.NeedConsensus())
		require.Equal(t, step2AO, out.NeedConsensus().BaseAliasOutput)
		require.Equal(t, uint32(2), out.NeedConsensus().LogIndex.AsUint32())
		require.Equal(t, cmtAddrA, &out.NeedConsensus().CommitteeAddr)
	}
	//
	// Make external committee rotation.
	rotateAO, _ := tcl.FakeRotationTX(step2AO, cmtAddrB)
	for nid := range nodes {
		tc.WithInput(nid, chainmanager.NewInputAliasOutputConfirmed(rotateAO))
	}
	tc.RunAll()
	tc.PrintAllStatusStrings("After external rotation", t.Logf)
	for _, n := range nodes {
		out := n.Output().(*chainmanager.Output)
		require.Equal(t, 0, out.NeedPublishTX().Size())
		require.NotNil(t, out.NeedConsensus())
		require.Equal(t, rotateAO, out.NeedConsensus().BaseAliasOutput)
		require.Equal(t, uint32(1), out.NeedConsensus().LogIndex.AsUint32())
		require.Equal(t, cmtAddrB, &out.NeedConsensus().CommitteeAddr)
	}
}
