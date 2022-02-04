// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	"time"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"
)

type MockedNode struct {
	PubKey          *ed25519.PublicKey
	Env             *MockedEnv
	store           kvstore.KVStore
	NodeConn        *testchain.MockedNodeConn
	ChainCore       *testchain.MockedChainCore
	ChainPeers      peering.PeerDomainProvider
	stateSync       coreutil.ChainStateSync
	Peers           peering.PeerDomainProvider
	StateManager    chain.StateManager
	StateTransition *testchain.MockedStateTransition
	Log             *logger.Logger
}

type MockedStateManagerMetrics struct{}

func (c *MockedStateManagerMetrics) RecordBlockSize(_ uint32, _ float64) {}

func NewMockedNode(env *MockedEnv, nodeIndex int, timers StateManagerTimers) *MockedNode {
	nodePubKey := env.NodePubKeys[nodeIndex]
	nodePubKeyStr := nodePubKey.String()[0:10]
	log := env.Log.Named(nodePubKeyStr)
	var peeringID peering.PeeringID
	copy(peeringID[:], env.ChainID[:iotago.AliasIDLength])
	peers, err := env.NetworkProviders[nodeIndex].PeerDomain(peeringID, env.NodePubKeys)
	require.NoError(env.T, err)
	stateMgrDomain, err := NewDomainWithFallback(peeringID, env.NetworkProviders[nodeIndex], log)
	require.NoError(env.T, err)
	ret := &MockedNode{
		PubKey:     nodePubKey,
		Env:        env,
		NodeConn:   testchain.NewMockedNodeConnection("Node_"+nodePubKeyStr, env.Ledger),
		store:      mapdb.NewMapDB(),
		stateSync:  coreutil.NewChainStateSync(),
		ChainCore:  testchain.NewMockedChainCore(env.T, env.ChainID, log),
		ChainPeers: peers,
		Peers:      peers,
		Log:        log,
	}

	stateMgrMetrics := new(MockedStateManagerMetrics)
	ret.ChainCore.OnGlobalStateSync(func() coreutil.ChainStateSync {
		return ret.stateSync
	})
	ret.ChainCore.OnGetStateReader(func() state.OptimisticStateReader {
		return state.NewOptimisticStateReader(ret.store, ret.stateSync)
	})
	/*ret.ChainPeers.Attach(peering.PeerMessageReceiverStateManager, func(peerMsg *peering.PeerMessageIn) {
		log.Debugf("State manager recvEvent from %v of type %v", peerMsg.SenderPubKey.String(), peerMsg.MsgType)
		switch peerMsg.MsgType {
		case peerMsgTypeGetBlock:
			msg, err := messages.NewGetBlockMsg(peerMsg.MsgData)
			if err != nil {
				log.Error(err)
				return
			}
			ret.StateManager.EnqueueGetBlockMsg(&messages.GetBlockMsgIn{
				GetBlockMsg:  *msg,
				SenderPubKey: peerMsg.SenderPubKey,
			})
		case peerMsgTypeBlock:
			msg, err := messages.NewBlockMsg(peerMsg.MsgData)
			if err != nil {
				log.Error(err)
				return
			}
			ret.StateManager.EnqueueBlockMsg(&messages.BlockMsgIn{
				BlockMsg:     *msg,
				SenderPubKey: peerMsg.SenderPubKey,
			})
		}
	})*/
	ret.StateManager = New(ret.store, ret.ChainCore, stateMgrDomain, ret.NodeConn, stateMgrMetrics, timers)
	/*ret.StateTransition = testchain.NewMockedStateTransition(env.T, env.StateKeyPair)
	ret.StateTransition.OnNextState(func(vstate state.VirtualStateAccess, tx *iotago.Transaction) {
		log.Debugf("MockedEnv.onNextState: state index %d", vstate.BlockIndex())
		stateOutput, err := utxoutil.GetSingleChainedAliasOutput(tx)
		require.NoError(env.T, err)
		go ret.StateManager.EnqueueStateCandidateMsg(vstate, stateOutput.ID())
		go ret.NodeConn.PostTransaction(tx)
	})
	ret.NodeConn.OnPostTransaction(func(tx *iotago.Transaction) {
		txID, err := tx.ID()
		require.NoError(env.T, err)
		log.Debugf("MockedNode.OnPostTransaction: transaction %v posted", txID)
		env.PostTransactionToLedger(tx)
	})
	ret.NodeConn.OnPullState(func() {
		log.Debugf("MockedNode.OnPullState request received")
		response := env.PullStateFromLedger()
		log.Debugf("MockedNode.OnPullState call EventStateMsg: chain output %s", iscp.OID(response.ChainOutput.ID()))
		go ret.StateManager.EnqueueStateMsg(response)
	})
	ret.NodeConn.OnPullConfirmedOutput(func(outputID *iotago.UTXOInput) {
		log.Debugf("MockedNode.OnPullConfirmedOutput %v", iscp.OID(outputID))
		response := env.PullConfirmedOutputFromLedger(outputID)
		log.Debugf("MockedNode.OnPullConfirmedOutput call EventOutputMsg")
		go ret.StateManager.EnqueueOutputMsg(response, outputID.ID())
	})*/
	ret.NodeConn.AttachToUnspentAliasOutputReceived(func(chainOutput *iscp.AliasOutputWithID, timestamp time.Time) {
		ret.StateManager.EnqueueStateMsg(&messages.StateMsg{
			ChainOutput: chainOutput,
			Timestamp:   timestamp,
		})
	})
	ret.NodeConn.AttachToTransactionReceived(func(*iotago.Transaction) {
		//TODO
		//ret.StateManager.EnqueueStateMsg
	})
	return ret
}

func (node *MockedNode) StartTimer() {
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
		chain.LogStateTransition(msg, nil, node.Log)
		if msg.ChainOutput.GetStateIndex() < limit {
			go node.StateTransition.NextState(msg.VirtualState, msg.ChainOutput, time.Now())
		}
	})
}

func (node *MockedNode) OnStateTransitionDoNothing() {
	node.ChainCore.OnStateTransition(func(msg *chain.ChainTransitionEventData) {})
}

func (node *MockedNode) MakeNewStateTransition() {
	node.StateTransition.NextState(node.StateManager.(*stateManager).solidState, node.StateManager.(*stateManager).stateOutput, time.Now())
}
