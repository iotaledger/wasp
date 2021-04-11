package statemgr

import (
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
	"sync"
	"testing"
)

type MockedEnv struct {
	T                 *testing.T
	Log               *logger.Logger
	Ledger            *utxodb.UtxoDB
	OriginatorKeyPair *ed25519.KeyPair
	OriginatorAddress ledgerstate.Address
	Peers             chain.PeerGroupProvider
	NodeConn          *mock_chain.MockedNodeConn
	ChainID           coretypes.ChainID
	mutex             sync.Mutex
	Nodes             map[string]*MockedNode
}

type MockedNode struct {
	Name            string
	Env             *MockedEnv
	Db              *dbprovider.DBProvider
	ChainCore       *mock_chain.MockedChainCore
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
		Peers:             mock_chain.NewDummyPeerGroup(),
		NodeConn:          mock_chain.NewMockedNodeConnection(),
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
	err = txBuilder.AddReminderOutputIfNeeded(ret.OriginatorAddress, nil)
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
		ret.Log.Infof("MockedEnv: posted transaction to ledger: %s", tx.ID().Base58())
	})
	ret.NodeConn.OnPullBacklog(func(addr ledgerstate.Address) {
		outputs := ret.Ledger.GetAddressOutputs(addr)
		require.EqualValues(t, 1, len(outputs))
		outTx, ok := ret.Ledger.GetTransaction(outputs[0].ID().TransactionID())
		require.True(t, ok)
		stateOutput, err := utxoutil.GetSingleChainedAliasOutput(outTx)
		require.NoError(t, err)

		ret.mutex.Lock()
		defer ret.mutex.Unlock()

		for _, node := range ret.Nodes {
			go func(manager chain.StateManager, log *logger.Logger) {
				log.Infof("MockedNodeConn.OnPullBacklog: call EventStateMsg: chain output %s", coretypes.OID(stateOutput.ID()))
				manager.EventStateMsg(&chain.StateMsg{
					ChainOutput: stateOutput,
					Timestamp:   outTx.Essence().Timestamp(),
				})
			}(node.StateManager, node.Log)
		}
	})

	return ret, originTx
}

func (env *MockedEnv) NewMockedNode(name string) *MockedNode {
	env.mutex.Lock()
	defer env.mutex.Unlock()

	log := env.Log.Named(name)
	ret := &MockedNode{
		Name:      name,
		Env:       env,
		Db:        dbprovider.NewInMemoryDBProvider(log),
		ChainCore: mock_chain.NewMockedChainCore(env.ChainID, log),
		Log:       log,
	}
	ret.StateManager = New(ret.Db, ret.ChainCore, env.Peers, env.NodeConn, log)
	ret.StateTransition = mock_chain.NewMockedStateTransition(env.T, env.Ledger, env.OriginatorKeyPair)
	ret.StateTransition.OnNextState(func(block state.Block, tx *ledgerstate.Transaction) {
		go ret.StateManager.EventBlockCandidateMsg(chain.BlockCandidateMsg{Block: block})
		go env.NodeConn.PostTransaction(tx, ret.ChainCore.ID().AsAddress(), 0)
	})
	return ret
}

func (env *MockedEnv) AddNode(node *MockedNode) {
	env.mutex.Lock()
	defer env.mutex.Unlock()
	env.Nodes[node.Name] = node
}

func (env *MockedEnv) RemoveNode(name string) {
	env.mutex.Lock()
	defer env.mutex.Unlock()
	delete(env.Nodes, name)
}
