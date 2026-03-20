// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainmanager_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/v2/packages/chain/chainmanager"
	"github.com/iotaledger/wasp/v2/packages/chain/committeelog"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/isc/isctest"
	"github.com/iotaledger/wasp/v2/packages/kvstore/mapdb"
	"github.com/iotaledger/wasp/v2/packages/origin"
	"github.com/iotaledger/wasp/v2/packages/parameters/parameterstest"
	"github.com/iotaledger/wasp/v2/packages/state"
	"github.com/iotaledger/wasp/v2/packages/state/statetest"
	"github.com/iotaledger/wasp/v2/packages/tcrypto"
	"github.com/iotaledger/wasp/v2/packages/testutil"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
	"github.com/iotaledger/wasp/v2/packages/testutil/testchain"
	"github.com/iotaledger/wasp/v2/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/v2/packages/testutil/testpeers"
)

func TestChainMgrBasic(t *testing.T) {
	type test struct {
		n int
		f int
	}
	tests := []test{
		{n: 1, f: 0},   // Low N.
		{n: 2, f: 0},   // Low N.
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
	defer log.Shutdown()
	//
	// Node identities and DKG.
	_, peerIdentities := testpeers.SetupKeys(uint16(n))
	nodeIDs := make([]gpa.NodeID, len(peerIdentities))
	for i, pid := range peerIdentities {
		nodeIDs[i] = gpa.NodeIDFromPublicKey(pid.GetPublicKey())
	}
	committeeAddrA, dkRegs := testpeers.SetupDistributedKeyGenerationTrivial(t, n, f, peerIdentities, nil)
	committeeAddrB, dkRegs := testpeers.SetupDistributedKeyGenerationTrivial(t, n, f, peerIdentities, dkRegs)
	require.NotNil(t, committeeAddrA)
	require.NotNil(t, committeeAddrB)
	t.Logf("Committee addressA: %v", committeeAddrA)
	t.Logf("Committee addressB: %v", committeeAddrB)
	//
	// Chain identifiers.
	committeeAddrASigner := testpeers.NewTestDistributedSignatureSigner(committeeAddrA, dkRegs, nodeIDs, peerIdentities, log)
	tcl := newTestChainLedger(t, committeeAddrASigner)
	anchor, deposit := tcl.MakeTxChainOrigin()
	//
	// Construct the nodes.
	nodes := map[gpa.NodeID]gpa.GPA{}
	stores := map[gpa.NodeID]state.Store{}
	needCons := map[gpa.NodeID]*chainmanager.NeedConsensusMap{}
	for i, nid := range nodeIDs {
		consensusStateRegistry := testutil.NewConsensusStateRegistry()
		stores[nid] = statetest.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB())
		_, err := origin.InitChainByStateMetadataBytes(stores[nid], anchor.GetStateMetadata(), deposit, parameterstest.L1Mock)
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
		needConsensusCB := func(upd *chainmanager.NeedConsensusMap) {
			needCons[nid] = upd
		}
		needPublishCB := func(upd *chainmanager.NeedPublishTXMap) {}
		cm, err := chainmanager.New(
			nid,
			anchor.ChainID(),
			stores[nid],
			consensusStateRegistry,
			dkRegs[i],
			gpa.NodeIDFromPublicKey,
			needConsensusCB,
			needPublishCB,
			activeAccessNodesCB,
			trackActiveStateCB,
			savePreliminaryBlockCB,
			updateCommitteeNodesCB,
			true, // deriveAnchorByQuorum
			-1,   // pipeliningLimit
			1,    // postponeRecoveryMilestones
			nil,  // metrics
			log.NewChildLogger(nid.ShortString()),
		)
		require.NoError(t, err)
		nodes[nid] = cm.AsGPA()
	}
	tc := gpa.NewTestContext(nodes)
	tc.PrintAllStatusStrings("Started", t.Logf)
	//
	// Provide initial Anchor.
	// Nevertheless, the first round after a reboot should have ⊥ as input to synchronize with each other.
	initAnchorInputs := map[gpa.NodeID]gpa.Input{}
	for nid := range nodes {
		initAnchorInputs[nid] = chainmanager.NewInputAnchorConfirmed(committeeAddrA, anchor)
	}
	tc.WithInputs(initAnchorInputs).RunAll()
	tc.PrintAllStatusStrings("Initial Anchor received", t.Logf)
	initAnchorLogIndex := committeelog.NilLogIndex()
	for nid, n := range nodes {
		out := n.Output().(*chainmanager.Output)
		ncm := needCons[nid]
		require.Equal(t, 0, out.NeedPublishTX().Size())
		require.NotNil(t, ncm)
		require.Equal(t, 1, ncm.Size())

		t.Logf("NeedConsensusMap after initial Anchor received: %v", ncm.AsMap())

		ncm.ForEach(func(nck chainmanager.NeedConsensusKey, nc *chainmanager.NeedConsensus) bool {
			require.Nil(t, nc.BaseStateAnchor)
			require.Equal(t, uint32(1), nc.LogIndex.AsUint32())
			require.Equal(t, committeeAddrA, &nc.CommitteeAddr)
			initAnchorLogIndex = nc.LogIndex
			return true
		})
	}
	//
	// All proposed NIL, thus consensus should output NIL as well.
	// So, we report consensus output to the chainMgr as ⊥.
	inputs := map[gpa.NodeID]gpa.Input{}
	for nid := range nodes {
		inputs[nid] = chainmanager.NewInputConsensusOutputSkip(*committeeAddrA, initAnchorLogIndex)
	}
	tc.WithInputs(inputs).RunAll()
	tc.PrintAllStatusStrings("Next Anchor received", t.Logf)

	// Checking that we did not immediately advance to the next LI, but instead scheduled it for the next Tick after the configured consensusDelay.
	for nid := range nodes {
		ncm := needCons[nid]
		require.Equal(t, 1, ncm.Size())
	}

	// Ticks are not sent automatically in this test - we wait just to verify it.
	time.Sleep(time.Second)
	for nid := range nodes {
		ncm := needCons[nid]
		require.Equal(t, 1, ncm.Size())
	}

	// Simulating tick
	inputs = map[gpa.NodeID]gpa.Input{}
	for nid := range nodes {
		inputs[nid] = chainmanager.NewInputCanPropose()
	}
	tc.WithInputs(inputs).RunAll()
	tc.PrintAllStatusStrings("CanPropose tick received", t.Logf)

	//
	// Now the next consensus instance should be requested.
	// Since the previous consensus decided ⊥, now all the nodes will propose the latest Anchor received from L1.
	for nid, n := range nodes {
		out := n.Output().(*chainmanager.Output)
		ncm := needCons[nid]
		require.Equal(t, 0, out.NeedPublishTX().Size())
		require.NotNil(t, ncm)
		t.Logf("NeedConsensusMap after next Anchor received: %v", ncm.AsMap())

		require.Equal(t, 2, ncm.Size())

		ncm.ForEach(func(nck chainmanager.NeedConsensusKey, nc *chainmanager.NeedConsensus) bool {
			switch nc.LogIndex.AsUint32() {
			case 1:
				require.Nil(t, nc.BaseStateAnchor)
				require.Equal(t, committeeAddrA, &nc.CommitteeAddr)
			case 2:
				require.Equal(t, anchor, nc.BaseStateAnchor)
				require.Equal(t, committeeAddrA, &nc.CommitteeAddr)
			default:
				panic("unexpected LI here")
			}
			return true
		})
	}
	//
	// Now we model the situation where the consensus for LI=2 produced a TX,
	// it was posted to the L1 and now we sending a response to the chain manager.
	tx1Digest := iotatest.RandomDigest()
	tx1OutSI := anchor.Anchor().Object.StateIndex + uint32(1)
	tx1OutAnchor := isctest.RandomStateAnchor(isctest.RandomAnchorOption{
		ID:         anchor.GetObjectID(),
		StateIndex: &tx1OutSI,
	})
	tc.WithInputs(lo.SliceToMap(nodeIDs, func(nid gpa.NodeID) (gpa.NodeID, gpa.Input) {
		return nid, chainmanager.NewInputChainTxPublishResult(
			*committeeAddrA,
			committeelog.LogIndex(2),
			*tx1Digest,
			&tx1OutAnchor,
			true,
		)
	})).RunAll()
	for nid, n := range nodes {
		out := n.Output().(*chainmanager.Output)
		ncm := needCons[nid]
		require.Equal(t, 0, out.NeedPublishTX().Size())
		require.NotNil(t, ncm)
		require.Equal(t, 3, ncm.Size())
		ncm.ForEach(func(nck chainmanager.NeedConsensusKey, nc *chainmanager.NeedConsensus) bool {
			switch nc.LogIndex.AsUint32() {
			case 1:
				require.Nil(t, nc.BaseStateAnchor)
				require.Equal(t, committeeAddrA, &nc.CommitteeAddr)
			case 2:
				require.Equal(t, anchor, nc.BaseStateAnchor)
				require.Equal(t, committeeAddrA, &nc.CommitteeAddr)
			case 3:
				require.Equal(t, &tx1OutAnchor, nc.BaseStateAnchor)
				require.Equal(t, committeeAddrA, &nc.CommitteeAddr)
			default:
				panic("unexpected LI here")
			}
			return true
		})
	}
}

func setupChainMgr(t *testing.T, n, f int) ( //nolint:gocritic
	[]gpa.NodeID,
	map[gpa.NodeID]gpa.GPA,
	map[gpa.NodeID]*chainmanager.NeedConsensusMap,
	*cryptolib.Address,
	*isc.StateAnchor,
	*gpa.TestContext,
) {
	t.Helper()
	log := testlogger.NewLogger(t)
	t.Cleanup(func() { log.Shutdown() })

	_, peerIdentities := testpeers.SetupKeys(uint16(n))
	nodeIDs := make([]gpa.NodeID, len(peerIdentities))
	for i, pid := range peerIdentities {
		nodeIDs[i] = gpa.NodeIDFromPublicKey(pid.GetPublicKey())
	}
	committeeAddr, dkRegs := testpeers.SetupDistributedKeyGenerationTrivial(t, n, f, peerIdentities, nil)
	require.NotNil(t, committeeAddr)

	committeeAddrSigner := testpeers.NewTestDistributedSignatureSigner(committeeAddr, dkRegs, nodeIDs, peerIdentities, log)
	tcl := newTestChainLedger(t, committeeAddrSigner)
	anchor, deposit := tcl.MakeTxChainOrigin()

	nodes := map[gpa.NodeID]gpa.GPA{}
	stores := map[gpa.NodeID]state.Store{}
	needCons := map[gpa.NodeID]*chainmanager.NeedConsensusMap{}
	for i, nid := range nodeIDs {
		consensusStateRegistry := testutil.NewConsensusStateRegistry()
		stores[nid] = statetest.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB())
		_, err := origin.InitChainByStateMetadataBytes(stores[nid], anchor.GetStateMetadata(), deposit, parameterstest.L1Mock)
		require.NoError(t, err)
		needConsensusCB := func(upd *chainmanager.NeedConsensusMap) {
			needCons[nid] = upd
		}
		cm, err := chainmanager.New(
			nid,
			anchor.ChainID(),
			stores[nid],
			consensusStateRegistry,
			dkRegs[i],
			gpa.NodeIDFromPublicKey,
			needConsensusCB,
			func(upd *chainmanager.NeedPublishTXMap) {},
			func() ([]*cryptolib.PublicKey, []*cryptolib.PublicKey) {
				return []*cryptolib.PublicKey{}, []*cryptolib.PublicKey{}
			},
			func(ao *isc.StateAnchor) {},
			func(state.Block) {},
			func(tcrypto.DKShare) {},
			true, // deriveAnchorByQuorum
			-1,   // pipeliningLimit
			1,    // postponeRecoveryMilestones
			nil,  // metrics
			log.NewChildLogger(nid.ShortString()),
		)
		require.NoError(t, err)
		nodes[nid] = cm.AsGPA()
	}
	tc := gpa.NewTestContext(nodes)
	return nodeIDs, nodes, needCons, committeeAddr, anchor, tc
}

// provideInitialAnchor sends an initial anchor to all nodes and returns the initial log index.
func provideInitialAnchor(
	t *testing.T,
	nodes map[gpa.NodeID]gpa.GPA,
	needCons map[gpa.NodeID]*chainmanager.NeedConsensusMap,
	committeeAddr *cryptolib.Address,
	anchor *isc.StateAnchor,
	tc *gpa.TestContext,
) committeelog.LogIndex {
	t.Helper()
	inputs := map[gpa.NodeID]gpa.Input{}
	for nid := range nodes {
		inputs[nid] = chainmanager.NewInputAnchorConfirmed(committeeAddr, anchor)
	}
	tc.WithInputs(inputs).RunAll()

	var initLI committeelog.LogIndex
	for nid := range nodes {
		ncm := needCons[nid]
		require.Equal(t, 1, ncm.Size())
		ncm.ForEach(func(_ chainmanager.NeedConsensusKey, nc *chainmanager.NeedConsensus) bool {
			initLI = nc.LogIndex
			return true
		})
	}
	return initLI
}

// sendSkip sends a consensus skip output to all nodes.
func sendSkip(
	nodes map[gpa.NodeID]gpa.GPA,
	committeeAddr *cryptolib.Address,
	li committeelog.LogIndex,
	tc *gpa.TestContext,
) {
	inputs := map[gpa.NodeID]gpa.Input{}
	for nid := range nodes {
		inputs[nid] = chainmanager.NewInputConsensusOutputSkip(*committeeAddr, li)
	}
	tc.WithInputs(inputs).RunAll()
}

// sendTick sends a CanPropose tick to all nodes.
func sendTick(
	nodes map[gpa.NodeID]gpa.GPA,
	tc *gpa.TestContext,
) {
	inputs := map[gpa.NodeID]gpa.Input{}
	for nid := range nodes {
		inputs[nid] = chainmanager.NewInputCanPropose()
	}
	tc.WithInputs(inputs).RunAll()
}

// maxLIInNeedConsensus returns the highest LogIndex in the NeedConsensusMap.
func maxLIInNeedConsensus(ncm *chainmanager.NeedConsensusMap) committeelog.LogIndex {
	var maxLI committeelog.LogIndex
	ncm.ForEach(func(_ chainmanager.NeedConsensusKey, nc *chainmanager.NeedConsensus) bool {
		if nc.LogIndex > maxLI {
			maxLI = nc.LogIndex
		}
		return true
	})
	return maxLI
}

// TestChainMgrConsecutiveSkips verifies that multiple consecutive skip decisions
// each require a tick before the next consensus instance is scheduled.
func TestChainMgrConsecutiveSkips(t *testing.T) {
	nodeIDs, nodes, needCons, committeeAddr, anchor, tc := setupChainMgr(t, 4, 1)
	initLI := provideInitialAnchor(t, nodes, needCons, committeeAddr, anchor, tc)

	for nid := range nodes {
		require.Equal(t, 1, needCons[nid].Size())
	}

	// First skip: should NOT immediately advance.
	sendSkip(nodes, committeeAddr, initLI, tc)
	for nid := range nodes {
		require.Equal(t, 1, needCons[nid].Size(), "skip should not immediately advance")
	}

	// Tick resolves the first skip → LI advances.
	sendTick(nodes, tc)
	for nid := range nodes {
		require.Equal(t, 2, needCons[nid].Size(), "tick should have advanced to next LI")
	}
	li2 := maxLIInNeedConsensus(needCons[nodeIDs[0]])

	// Second skip: should NOT immediately advance.
	sendSkip(nodes, committeeAddr, li2, tc)
	for nid := range nodes {
		require.Equal(t, 2, needCons[nid].Size(), "second skip should not immediately advance")
	}

	// Second tick resolves the second skip.
	sendTick(nodes, tc)
	for nid := range nodes {
		require.Equal(t, 3, needCons[nid].Size(), "second tick should advance to next LI")
	}
	li3 := maxLIInNeedConsensus(needCons[nodeIDs[0]])
	require.Greater(t, li3.AsUint32(), li2.AsUint32())
}

// TestChainMgrSkipThenDoneBeforeTick verifies that if a consensus Done
// (publish result) arrives before the tick resolves a pending skip,
// the system behaves correctly — the Done advances the log index, and
// the subsequent tick does not cause a spurious double-advance.
func TestChainMgrSkipThenDoneBeforeTick(t *testing.T) {
	nodeIDs, nodes, needCons, committeeAddr, anchor, tc := setupChainMgr(t, 4, 1)
	initLI := provideInitialAnchor(t, nodes, needCons, committeeAddr, anchor, tc)

	// Skip at initLI: pending, not yet advanced.
	sendSkip(nodes, committeeAddr, initLI, tc)
	for nid := range nodes {
		require.Equal(t, 1, needCons[nid].Size())
	}

	// Before the tick, simulate a publish result (Done) at initLI.
	// This advances the log index immediately.
	txDigest := iotatest.RandomDigest()
	nextSI := anchor.Anchor().Object.StateIndex + uint32(1)
	nextAnchor := isctest.RandomStateAnchor(isctest.RandomAnchorOption{
		ID:         anchor.GetObjectID(),
		StateIndex: &nextSI,
	})
	doneInputs := map[gpa.NodeID]gpa.Input{}
	for _, nid := range nodeIDs {
		doneInputs[nid] = chainmanager.NewInputChainTxPublishResult(
			*committeeAddr, initLI, *txDigest, &nextAnchor, true,
		)
	}
	tc.WithInputs(doneInputs).RunAll()

	// The Done should have advanced the log index.
	sizeAfterDone := needCons[nodeIDs[0]].Size()
	require.Greater(t, sizeAfterDone, 1, "Done should have advanced the log index")

	// Now send a tick. The pending skip should be a no-op (already superseded by Done).
	sendTick(nodes, tc)
	for nid := range nodes {
		require.Equal(t, sizeAfterDone, needCons[nid].Size(),
			"tick after Done should not cause spurious advance")
	}
}

func newTestChainLedger(t *testing.T, originator cryptolib.Signer) *testchain.TestChainLedger {
	l1client := l1starter.Instance().L1Client()
	l1client.RequestFunds(context.Background(), *originator.Address())
	l1client.RequestFunds(context.Background(), *originator.Address())

	iscPackage, err := l1client.L2().DeployISCContracts(context.Background(), cryptolib.SignerToIotaSigner(originator))
	require.NoError(t, err)

	return testchain.NewTestChainLedger(t, originator, &iscPackage, l1client)
}
