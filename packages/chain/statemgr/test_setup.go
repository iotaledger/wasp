package statemgr

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/peering"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxodb"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/dbprovider"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
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
	ChainID           coretypes.ChainID
	mutex             sync.Mutex
	Nodes             map[string]*MockedNode
	push              bool
}

type MockedNode struct {
	NetID           string
	Env             *MockedEnv
	Db              *dbprovider.DBProvider
	NodeConn        *testchain.MockedNodeConn
	ChainCore       *testchain.MockedChainCore
	Peers           *testchain.MockedPeerDomainProvider
	StateManager    chain.StateManager
	StateTransition *testchain.MockedStateTransition
	Log             *logger.Logger
}

func NewMockedEnv(t *testing.T, debug bool) (*MockedEnv, *ledgerstate.Transaction) {
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
		Nodes:             make(map[string]*MockedNode),
	}
	ret.OriginatorKeyPair, ret.OriginatorAddress = ret.Ledger.NewKeyPairByIndex(0)
	_, err := ret.Ledger.RequestFunds(ret.OriginatorAddress)
	require.NoError(t, err)

	outputs := ret.Ledger.GetAddressOutputs(ret.OriginatorAddress)
	require.True(t, len(outputs) == 1)

	bals := map[ledgerstate.Color]uint64{ledgerstate.ColorIOTA: 100}
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

	ret.ChainID = *coretypes.NewChainID(retOut.GetAliasAddress())

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
		go node.StateManager.EventStateMsg(&chain.StateMsg{
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

func (env *MockedEnv) PullStateFromLedger(addr *ledgerstate.AliasAddress) *chain.StateMsg {
	env.Log.Debugf("MockedEnv.PullStateFromLedger request received for address %v", addr.Base58)
	outputs := env.Ledger.GetAddressOutputs(addr)
	require.EqualValues(env.T, 1, len(outputs))
	outTx, ok := env.Ledger.GetTransaction(outputs[0].ID().TransactionID())
	require.True(env.T, ok)
	stateOutput, err := utxoutil.GetSingleChainedAliasOutput(outTx)
	require.NoError(env.T, err)

	env.Log.Debugf("MockedEnv.PullStateFromLedger chain output %s found", coretypes.OID(stateOutput.ID()))
	return &chain.StateMsg{
		ChainOutput: stateOutput,
		Timestamp:   outTx.Essence().Timestamp(),
	}
}

func (env *MockedEnv) PullConfirmedOutputFromLedger(addr ledgerstate.Address, outputID ledgerstate.OutputID) ledgerstate.Output {
	env.Log.Debugf("MockedEnv.PullConfirmedOutputFromLedger for address %v output %v", addr.Base58, coretypes.OID(outputID))
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

func (env *MockedEnv) NewMockedNode(netid string, allPeers []string, timers Timers) *MockedNode {
	log := env.Log.Named(netid)
	ret := &MockedNode{
		NetID:     netid,
		Env:       env,
		NodeConn:  testchain.NewMockedNodeConnection(),
		Db:        dbprovider.NewInMemoryDBProvider(log),
		ChainCore: testchain.NewMockedChainCore(env.ChainID, log),
		Peers:     testchain.NewMockedPeerDomain(netid, allPeers, log),
		Log:       log,
	}
	ret.StateManager = New(ret.Db, ret.ChainCore, ret.Peers, ret.NodeConn, log, timers)
	ret.StateTransition = testchain.NewMockedStateTransition(env.T, env.OriginatorKeyPair)
	ret.StateTransition.OnNextState(func(vstate state.VirtualState, tx *ledgerstate.Transaction) {
		log.Debugf("MockedEnv.onNextState: state index %d", vstate.BlockIndex())
		go ret.StateManager.EventStateCandidateMsg(chain.StateCandidateMsg{State: vstate})
		go ret.NodeConn.PostTransaction(tx)
	})
	ret.NodeConn.OnPostTransaction(func(tx *ledgerstate.Transaction) {
		log.Debugf("MockedNode.OnPostTransaction: transaction %v posted", tx.ID().Base58())
		env.PostTransactionToLedger(tx)
	})
	ret.NodeConn.OnPullState(func(addr *ledgerstate.AliasAddress) {
		log.Debugf("MockedNode.OnPullState request received for address %v", addr.Base58)
		response := env.PullStateFromLedger(addr)
		log.Debugf("MockedNode.OnPullState call EventStateMsg: chain output %s", coretypes.OID(response.ChainOutput.ID()))
		go ret.StateManager.EventStateMsg(response)
	})
	ret.NodeConn.OnPullConfirmedOutput(func(addr ledgerstate.Address, outputID ledgerstate.OutputID) {
		log.Debugf("MockedNode.OnPullConfirmedOutput %v", coretypes.OID(outputID))
		response := env.PullConfirmedOutputFromLedger(addr, outputID)
		log.Debugf("MockedNode.OnPullConfirmedOutput call EventOutputMsg")
		go ret.StateManager.EventOutputMsg(response)
	})

	return ret
}

func (node *MockedNode) StartTimer() {
	go func() {
		node.StateManager.Ready().MustWait()
		counter := 0
		for {
			node.StateManager.EventTimerMsg(chain.TimerTick(counter))
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
		syncInfo = node.StateManager.GetSyncInfo()
		if syncInfo != nil && syncInfo.SyncedBlockIndex >= index {
			return syncInfo, nil
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (env *MockedEnv) AddNode(node *MockedNode) {
	env.mutex.Lock()
	defer env.mutex.Unlock()

	if _, ok := env.Nodes[node.NetID]; ok {
		env.Log.Panicf("AddNode: duplicate node index %s", node.NetID)
	}
	env.Nodes[node.NetID] = node
}

func (env *MockedEnv) RemoveNode(netID string) {
	env.mutex.Lock()
	defer env.mutex.Unlock()
	delete(env.Nodes, netID)
}

// SetupPeerGroupSimple sets up simple communication between nodes
func (nT *MockedNode) SetupPeerGroupSimple() {
	nT.Peers.OnSend(func(target string, peerMsg *peering.PeerMessage) {
		nT.Log.Debugf("MockedNode:OnSendMsg to peer %s", target)
		node, ok := nT.Env.Nodes[target]
		if !ok {
			nT.Env.Log.Warnf("node %s: wrong target netID %s", nT.NetID, target)
			return
		}
		rdr := bytes.NewReader(peerMsg.MsgData)
		switch peerMsg.MsgType {
		case chain.MsgGetBlock:
			nT.Log.Debugf("MockedNode:OnSendMsg MsgGetBlock received")
			msg := chain.GetBlockMsg{}
			if err := msg.Read(rdr); err != nil {
				nT.Env.Log.Errorf("error reading MsgGetBlock message: %v", err)
				return
			}
			msg.SenderNetID = nT.NetID
			go node.StateManager.EventGetBlockMsg(&msg)

		case chain.MsgBlock:
			nT.Log.Debugf("MockedNode:OnSendMsg MsgBlock received")
			msg := chain.BlockMsg{}
			if err := msg.Read(rdr); err != nil {
				nT.Env.Log.Errorf("error reading MsgBlock message: %v", err)
				return
			}
			msg.SenderNetID = peerMsg.SenderNetID
			go node.StateManager.EventBlockMsg(&msg)

		default:
			nT.Log.Panicf("msg type %d not implemented in the simple mocked peer group")
		}
	})
}
