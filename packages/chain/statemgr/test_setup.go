package statemgr

import (
	"bytes"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxodb"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/mock_chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/dbprovider"
	"github.com/iotaledger/wasp/packages/state"
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
	NodeConn          *mock_chain.MockedNodeConn
	ChainID           coretypes.ChainID
	mutex             sync.Mutex
	Nodes             map[uint16]*MockedNode
	push              bool
}

type MockedNode struct {
	Index           uint16
	Env             *MockedEnv
	Db              *dbprovider.DBProvider
	ChainCore       *mock_chain.MockedChainCore
	Peers           *mock_chain.MockedPeerGroupProvider
	StateManager    chain.StateManager
	StateTransition *mock_chain.MockedStateTransition
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
		NodeConn:          mock_chain.NewMockedNodeConnection(),
		Nodes:             make(map[uint16]*MockedNode),
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

	ret.NodeConn.OnPostTransaction(func(tx *ledgerstate.Transaction, _ ledgerstate.Address, _ uint16) {
		_, exists := ret.Ledger.GetTransaction(tx.ID())
		if exists {
			ret.Log.Debugf("posted repeating originTx: %s", tx.ID().Base58())
			return
		}
		if err := ret.Ledger.AddTransaction(tx); err != nil {
			ret.Log.Error(err)
			return
		}
		// Push transaction to nodes
		go ret.pushStateToNodesIfSet(tx)

		ret.Log.Infof("MockedEnv: posted transaction to ledger: %s", tx.ID().Base58())
	})
	var pullStateOutputCounter uint64 = 0
	pullStateOutputClosure := func(addr *ledgerstate.AliasAddress) {
		requestID := atomic.AddUint64(&pullStateOutputCounter, 1)
		log.Infof("MockedNodeConn.OnPullBacklog request %d received for address %v", requestID, addr.Base58)
		outputs := ret.Ledger.GetAddressOutputs(addr)
		require.EqualValues(t, 1, len(outputs))
		outTx, ok := ret.Ledger.GetTransaction(outputs[0].ID().TransactionID())
		require.True(t, ok)
		stateOutput, err := utxoutil.GetSingleChainedAliasOutput(outTx)
		require.NoError(t, err)

		ret.mutex.Lock()
		defer ret.mutex.Unlock()

		// TODO: avoid broadcast
		for _, node := range ret.Nodes {
			go func(manager chain.StateManager, log *logger.Logger) {
				log.Infof("MockedNodeConn.OnPullBacklog request %d: call EventStateMsg: chain output %s", requestID, coretypes.OID(stateOutput.ID()))
				manager.EventStateMsg(&chain.StateMsg{
					ChainOutput: stateOutput,
					Timestamp:   outTx.Essence().Timestamp(),
				})
			}(node.StateManager, node.Log)
		}
	}
	ret.NodeConn.OnPullState(pullStateOutputClosure)

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

func (env *MockedEnv) NewMockedNode(index uint16) *MockedNode {
	log := env.Log.Named(fmt.Sprintf("n%d", index))
	ret := &MockedNode{
		Index:     index,
		Env:       env,
		Db:        dbprovider.NewInMemoryDBProvider(log),
		ChainCore: mock_chain.NewMockedChainCore(env.ChainID, log),
		Peers:     mock_chain.NewMockedPeerGroupProvider(),
		Log:       log,
	}
	ret.StateManager = New(ret.Db, ret.ChainCore, chain.NewPeerGroup(ret.Peers), env.NodeConn, log, Timers{}.SetPullStateRetry(10*time.Millisecond).SetPullStateNewBlockDelay(50*time.Millisecond))
	ret.StateTransition = mock_chain.NewMockedStateTransition(env.T, env.OriginatorKeyPair)
	ret.StateTransition.OnNextState(func(vstate state.VirtualState, tx *ledgerstate.Transaction) {
		ret.Log.Debugf("MockedEnv.OnNextState: state index %d", vstate.BlockIndex())
		go ret.StateManager.EventStateCandidateMsg(chain.StateCandidateMsg{State: vstate})
		go env.NodeConn.PostTransaction(tx, ret.ChainCore.ID().AsAddress(), 0)
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

	if _, ok := env.Nodes[node.Index]; ok {
		env.Log.Panicf("AddNode: duplicate node index %d", node.Index)
	}
	env.Nodes[node.Index] = node
}

func (env *MockedEnv) RemoveNode(index uint16) {
	env.mutex.Lock()
	defer env.mutex.Unlock()
	delete(env.Nodes, index)
}

// SetupPeerGroupSimple sets up simple communication between nodes
func (nT *MockedNode) SetupPeerGroupSimple() {
	nT.Peers.OnLock(func() {
		nT.Env.mutex.Lock()
	})
	nT.Peers.OnUnlock(func() {
		nT.Env.mutex.Unlock()
	})
	nT.Peers.OnNumPeers(func() uint16 {
		return uint16(len(nT.Env.Nodes))
	})
	nT.Peers.OnNumIsAlive(func(q uint16) bool {
		nT.Env.mutex.Lock()
		defer nT.Env.mutex.Unlock()
		return q <= uint16(len(nT.Env.Nodes))
	})
	nT.Peers.OnSendMsg(func(targetPeerIndex uint16, msgType byte, msgData []byte) error {
		nT.Log.Infof("XXX MockedPeerGroup:OnSendMsg peer %v", targetPeerIndex)
		node, ok := nT.Env.Nodes[targetPeerIndex]
		if !ok {
			return fmt.Errorf("wrong peer index %d", targetPeerIndex)
		}
		rdr := bytes.NewReader(msgData)
		switch msgType {
		case chain.MsgGetBlock:
			nT.Log.Infof("XXX MockedPeerGroup:OnSendMsg MsgGetBlock")
			msg := chain.GetBlockMsg{}
			if err := msg.Read(rdr); err != nil {
				return fmt.Errorf("error reading MsgGetBlock message: %v", err)
			}
			msg.SenderIndex = nT.Index
			go node.StateManager.EventGetBlockMsg(&msg)

		case chain.MsgBlock:
			nT.Log.Infof("XXX MockedPeerGroup:OnSendMsg MsgBlock")
			msg := chain.BlockMsg{}
			if err := msg.Read(rdr); err != nil {
				return fmt.Errorf("error reading MsgBlock message: %v", err)
			}
			go node.StateManager.EventBlockMsg(&msg)

		default:
			nT.Log.Panicf("msg type %d not implemented in the simple mocked peer group")
		}
		return nil
	})
}
