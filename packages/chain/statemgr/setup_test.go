// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	"io"
	"sync"
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/iscp/colored"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxodb"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"golang.org/x/xerrors"
)

type MockedEnv struct {
	T                 *testing.T
	Log               *logger.Logger
	Ledger            *utxodb.UtxoDB
	OriginatorKeyPair *ed25519.KeyPair
	OriginatorAddress ledgerstate.Address
	NodePubKeys       []*ed25519.PublicKey
	NetworkProviders  []peering.NetworkProvider
	NetworkBehaviour  *testutil.PeeringNetDynamic
	NetworkCloser     io.Closer
	ChainID           *iscp.ChainID
	mutex             sync.Mutex
	Nodes             map[ed25519.PublicKey]*MockedNode
	push              bool
}

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

func NewMockedEnv(nodeCount int, t *testing.T, debug bool) (*MockedEnv, *ledgerstate.Transaction) {
	level := zapcore.InfoLevel
	if debug {
		level = zapcore.DebugLevel
	}
	log := testlogger.WithLevel(testlogger.NewLogger(t, "04:05.000"), level, false)
	ret := &MockedEnv{
		T:                 t,
		Log:               log,
		Ledger:            utxodb.New(),
		OriginatorKeyPair: nil,
		OriginatorAddress: nil,
		Nodes:             make(map[ed25519.PublicKey]*MockedNode),
	}
	ret.OriginatorKeyPair, ret.OriginatorAddress = ret.Ledger.NewKeyPairByIndex(0)
	_, err := ret.Ledger.RequestFunds(ret.OriginatorAddress)
	require.NoError(t, err)

	outputs := ret.Ledger.GetAddressOutputs(ret.OriginatorAddress)
	require.True(t, len(outputs) == 1)

	bals := colored.ToL1Map(colored.NewBalancesForIotas(100))

	txBuilder := utxoutil.NewBuilder(outputs...)
	err = txBuilder.AddNewAliasMint(bals, ret.OriginatorAddress, state.OriginStateHash().Bytes())
	require.NoError(t, err)
	err = txBuilder.AddRemainderOutputIfNeeded(ret.OriginatorAddress, nil)
	require.NoError(t, err)
	originTx, err := txBuilder.BuildWithED25519(ret.OriginatorKeyPair)
	require.NoError(t, err)
	err = ret.Ledger.AddTransaction(originTx)
	require.NoError(t, err)

	retOut, err := utxoutil.GetSingleChainedAliasOutput(originTx)
	require.NoError(t, err)

	ret.ChainID = iscp.NewChainID(retOut.GetAliasAddress())

	ret.NetworkBehaviour = testutil.NewPeeringNetDynamic(log)

	nodeIDs, nodeIdentities := testpeers.SetupKeys(uint16(nodeCount))
	ret.NodePubKeys = testpeers.PublicKeys(nodeIdentities)
	ret.NetworkProviders, ret.NetworkCloser = testpeers.SetupNet(nodeIDs, nodeIdentities, ret.NetworkBehaviour, log)

	return ret, originTx
}

func (env *MockedEnv) SetPushStateToNodesOption(push bool) {
	env.mutex.Lock()
	defer env.mutex.Unlock()
	env.push = push
}

func (env *MockedEnv) pushStateToNodesIfSet(tx *ledgerstate.Transaction) {
	env.mutex.Lock()
	defer env.mutex.Unlock()

	if !env.push {
		return
	}
	stateOutput, err := utxoutil.GetSingleChainedAliasOutput(tx)
	require.NoError(env.T, err)

	for _, node := range env.Nodes {
		go node.StateManager.EnqueueStateMsg(&messages.StateMsg{
			ChainOutput: stateOutput,
			Timestamp:   tx.Essence().Timestamp(),
		})
	}
}

func (env *MockedEnv) PostTransactionToLedger(tx *ledgerstate.Transaction) {
	env.Log.Debugf("MockedEnv.PostTransactionToLedger: transaction %v", tx.ID().Base58())
	_, exists := env.Ledger.GetTransaction(tx.ID())
	if exists {
		env.Log.Debugf("MockedEnv.PostTransactionToLedger: posted repeating originTx: %s", tx.ID().Base58())
		return
	}
	if err := env.Ledger.AddTransaction(tx); err != nil {
		env.Log.Errorf("MockedEnv.PostTransactionToLedger: error adding transaction: %v", err)
		return
	}
	// Push transaction to nodes
	go env.pushStateToNodesIfSet(tx)

	env.Log.Infof("MockedEnv.PostTransactionToLedger: posted transaction to ledger: %s", tx.ID().Base58())
}

func (env *MockedEnv) PullStateFromLedger() *messages.StateMsg {
	env.Log.Debugf("MockedEnv.PullStateFromLedger request received")
	outputs := env.Ledger.GetAddressOutputs(env.ChainID.AsAliasAddress())
	require.EqualValues(env.T, 1, len(outputs))
	outTx, ok := env.Ledger.GetTransaction(outputs[0].ID().TransactionID())
	require.True(env.T, ok)
	stateOutput, err := utxoutil.GetSingleChainedAliasOutput(outTx)
	require.NoError(env.T, err)

	env.Log.Debugf("MockedEnv.PullStateFromLedger chain output %s found", iscp.OID(stateOutput.ID()))
	return &messages.StateMsg{
		ChainOutput: stateOutput,
		Timestamp:   outTx.Essence().Timestamp(),
	}
}

func (env *MockedEnv) PullConfirmedOutputFromLedger(outputID ledgerstate.OutputID) ledgerstate.Output {
	env.Log.Debugf("MockedEnv.PullConfirmedOutputFromLedger for output %v", iscp.OID(outputID))
	tx, foundTx := env.Ledger.GetTransaction(outputID.TransactionID())
	require.True(env.T, foundTx)
	outputIndex := outputID.OutputIndex()
	outputs := tx.Essence().Outputs()
	require.True(env.T, int(outputIndex) < len(outputs))
	output := outputs[outputIndex].UpdateMintingColor()
	require.NotNil(env.T, output)
	env.Log.Debugf("MockedEnv.PullConfirmedOutputFromLedger output found")
	return output
}

func (env *MockedEnv) NewMockedNode(nodeIndex int, timers StateManagerTimers) *MockedNode {
	nodePubKey := env.NodePubKeys[nodeIndex]
	nodePubKeyStr := nodePubKey.String()[0:10]
	log := env.Log.Named(nodePubKeyStr)
	peers, err := env.NetworkProviders[nodeIndex].PeerDomain(env.ChainID.Array(), env.NodePubKeys)
	require.NoError(env.T, err)
	stateMgrDomain, err := NewDomainWithFallback(env.ChainID.Array(), env.NetworkProviders[nodeIndex], log)
	require.NoError(env.T, err)
	ret := &MockedNode{
		PubKey:     nodePubKey,
		Env:        env,
		NodeConn:   testchain.NewMockedNodeConnection("Node_" + nodePubKeyStr),
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
	ret.ChainPeers.Attach(peering.PeerMessageReceiverStateManager, func(peerMsg *peering.PeerMessageIn) {
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
	})
	ret.StateManager = New(ret.store, ret.ChainCore, stateMgrDomain, ret.NodeConn, stateMgrMetrics, timers)
	ret.StateTransition = testchain.NewMockedStateTransition(env.T, env.OriginatorKeyPair)
	ret.StateTransition.OnNextState(func(vstate state.VirtualStateAccess, tx *ledgerstate.Transaction) {
		log.Debugf("MockedEnv.onNextState: state index %d", vstate.BlockIndex())
		stateOutput, err := utxoutil.GetSingleChainedAliasOutput(tx)
		require.NoError(env.T, err)
		go ret.StateManager.EnqueueStateCandidateMsg(vstate, stateOutput.ID())
		go ret.NodeConn.PostTransaction(tx)
	})
	ret.NodeConn.OnPostTransaction(func(tx *ledgerstate.Transaction) {
		log.Debugf("MockedNode.OnPostTransaction: transaction %v posted", tx.ID().Base58())
		env.PostTransactionToLedger(tx)
	})
	ret.NodeConn.OnPullState(func() {
		log.Debugf("MockedNode.OnPullState request received")
		response := env.PullStateFromLedger()
		log.Debugf("MockedNode.OnPullState call EventStateMsg: chain output %s", iscp.OID(response.ChainOutput.ID()))
		go ret.StateManager.EnqueueStateMsg(response)
	})
	ret.NodeConn.OnPullConfirmedOutput(func(outputID ledgerstate.OutputID) {
		log.Debugf("MockedNode.OnPullConfirmedOutput %v", iscp.OID(outputID))
		response := env.PullConfirmedOutputFromLedger(outputID)
		log.Debugf("MockedNode.OnPullConfirmedOutput call EventOutputMsg")
		go ret.StateManager.EnqueueOutputMsg(response)
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

func (env *MockedEnv) AddNode(node *MockedNode) {
	env.mutex.Lock()
	defer env.mutex.Unlock()

	if _, ok := env.Nodes[*node.PubKey]; ok {
		env.Log.Panicf("AddNode: duplicate node index %s", node.PubKey.String())
	}
	env.Nodes[*node.PubKey] = node
}

func (env *MockedEnv) RemoveNode(node *MockedNode) {
	env.mutex.Lock()
	defer env.mutex.Unlock()
	delete(env.Nodes, *node.PubKey)
}
