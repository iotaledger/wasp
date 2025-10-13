// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainmanager_test

import (
	"context"
	"fmt"
	"testing"

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
	//
	// Now the next consensus instance should be requested.
	// Since the previous consensus decided ⊥, now all the nodes will propose the latest Anchor received from L1.
	for nid, n := range nodes {
		out := n.Output().(*chainmanager.Output)
		ncm := needCons[nid]
		require.Equal(t, 0, out.NeedPublishTX().Size())
		require.NotNil(t, ncm)
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

func newTestChainLedger(t *testing.T, originator cryptolib.Signer) *testchain.TestChainLedger {
	l1client := l1starter.Instance().L1Client()
	l1client.RequestFunds(context.Background(), *originator.Address())
	l1client.RequestFunds(context.Background(), *originator.Address())

	iscPackage, err := l1client.DeployISCContracts(context.Background(), cryptolib.SignerToIotaSigner(originator))
	require.NoError(t, err)

	return testchain.NewTestChainLedger(t, originator, &iscPackage, l1client)
}
