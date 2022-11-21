// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/core/kvstore/mapdb"
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/inx-app/pkg/nodebridge"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	"github.com/iotaledger/wasp/packages/wal"
)

type MockedNode struct {
	PubKey       *cryptolib.PublicKey
	Env          *MockedEnv
	NodeConn     *testchain.MockedNodeConn
	ChainCore    *testchain.MockedChainCore
	StateManager chain.StateManager
	Log          *logger.Logger
}

type MockedStateManagerMetrics struct{}

func (c *MockedStateManagerMetrics) RecordBlockSize(_ uint32, _ float64) {}

func (c *MockedStateManagerMetrics) LastSeenStateIndex(_ uint32) {}

func NewMockedNode(env *MockedEnv, nodeIndex int, timers StateManagerTimers) *MockedNode {
	nodeID := env.NodeIDs[nodeIndex]
	log := env.Log.Named(nodeID)
	var peeringID peering.PeeringID
	copy(peeringID[:], env.ChainID[:iotago.AliasIDLength])
	stateMgrDomain, err := NewDomainWithFallback(peeringID, env.NetworkProviders[nodeIndex], log)
	require.NoError(env.T, err)
	ret := &MockedNode{
		PubKey:    env.NodePubKeys[nodeIndex],
		Env:       env,
		NodeConn:  testchain.NewMockedNodeConnection("Node_"+nodeID, env.Ledgers, log),
		ChainCore: testchain.NewMockedChainCore(env.T, env.ChainID, log),
		Log:       log,
	}

	stateSync := coreutil.NewChainStateSync()
	store := mapdb.NewMapDB()
	stateMgrMetrics := new(MockedStateManagerMetrics)
	ret.ChainCore.OnGlobalStateSync(func() coreutil.ChainStateSync {
		return stateSync
	})
	ret.ChainCore.OnGetStateReader(func() state.OptimisticStateReader {
		return state.NewOptimisticStateReader(store, stateSync)
	})
	ret.StateManager = New(store, ret.ChainCore, stateMgrDomain, ret.NodeConn, stateMgrMetrics, wal.NewDefault(), false, "", true, timers)
	ret.Log.Debugf("Mocked node %v created: id %v public key %v", nodeIndex, nodeID, ret.PubKey.String())

	ret.NodeConn.RegisterChain(
		env.ChainID,
		func(oid iotago.OutputID, o iotago.Output) {
			ret.StateManager.EnqueueAliasOutput(isc.NewAliasOutputWithID(o.(*iotago.AliasOutput), oid.UTXOInput()))
		},
		func(iotago.OutputID, iotago.Output) {},
		func(*nodebridge.Milestone) {},
	)

	return ret
}

func (node *MockedNode) Start() {
	// node.ChainNodeConn.AttachToAliasOutput(node.StateManager.EnqueueAliasOutput)
	node.startTimer()
	node.Log.Debugf("Mocked node %v started", node.PubKey.String())
}

func (node *MockedNode) startTimer() {
	go func() {
		node.StateManager.Ready().MustWait()
		counter := 0
		for {
			node.StateManager.EnqueueTimerMsg(messages.TimerTick(counter))
			counter++
			time.Sleep(50 * time.Millisecond)
		}
	}()
}

func (node *MockedNode) WaitSyncBlockIndex(index uint32, timeout time.Duration) (*chain.SyncInfo, error) {
	deadline := time.Now().Add(timeout)
	var syncInfo *chain.SyncInfo
	for {
		if time.Now().After(deadline) {
			return nil, xerrors.Errorf("WaitSyncBlockIndex: target index %d, timeout %v reached", index, timeout)
		}
		syncInfo = node.StateManager.GetStatusSnapshot()
		if syncInfo != nil && syncInfo.SyncedBlockIndex >= index {
			return syncInfo, nil
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (node *MockedNode) OnStateTransitionMakeNewStateTransition(limit uint32) {
	node.ChainCore.OnStateTransition(func(msg *chain.ChainTransitionEventData) {
		chain.LogStateTransition(msg.VirtualState.BlockIndex(), isc.OID(msg.ChainOutput.ID()), state.RootCommitment(msg.VirtualState.TrieNodeStore()), nil, node.Log)
		if msg.ChainOutput.GetStateIndex() < limit {
			go node.NextState(msg.VirtualState, msg.ChainOutput)
		}
	})
}

func (node *MockedNode) OnStateTransitionDoNothing() {
	node.ChainCore.OnStateTransition(func(msg *chain.ChainTransitionEventData) {})
}

func (node *MockedNode) MakeNewStateTransition() {
	node.NextState(node.StateManager.(*stateManager).solidState, node.StateManager.(*stateManager).stateOutput)
}

func (node *MockedNode) NextState(vstate state.VirtualStateAccess, chainOutput *isc.AliasOutputWithID) {
	node.Log.Debugf("NextState: from state %d, output ID %v", vstate.BlockIndex(), isc.OID(chainOutput.ID()))
	nextState, tx, aliasOutputID := testchain.NextState(node.Env.T, node.Env.StateKeyPair, vstate, chainOutput, time.Now())
	cid := isc.ChainIDFromAliasID(chainOutput.GetAliasID())
	go node.NodeConn.PublishTransaction(&cid, tx)
	go node.StateManager.EnqueueStateCandidateMsg(nextState, aliasOutputID)
	node.Log.Debugf("NextState: result state %d, output ID %v", nextState.BlockIndex(), isc.OID(aliasOutputID))
}
