// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/metrics"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxodb"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/committee"
	"github.com/iotaledger/wasp/packages/chain/mempool"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

type MockedEnv struct {
	T                 *testing.T
	Quorum            uint16
	Log               *logger.Logger
	Ledger            *utxodb.UtxoDB
	StateAddress      ledgerstate.Address
	OriginatorKeyPair *ed25519.KeyPair
	OriginatorAddress ledgerstate.Address
	NodeIDs           []string
	NetworkProviders  []peering.NetworkProvider
	NetworkBehaviour  *testutil.PeeringNetDynamic
	NetworkCloser     io.Closer
	DKSRegistries     []registry.DKShareRegistryProvider
	ChainID           iscp.ChainID
	MockedACS         chain.AsynchronousCommonSubsetRunner
	InitStateOutput   *ledgerstate.AliasOutput
	mutex             sync.Mutex
	Nodes             []*mockedNode
}

type mockedNode struct {
	NodeID      string
	Env         *MockedEnv
	NodeConn    *testchain.MockedNodeConn  // GoShimmer mock
	ChainCore   *testchain.MockedChainCore // Chain mock
	stateSync   coreutil.ChainStateSync    // Chain mock
	Mempool     chain.Mempool              // Consensus needs
	Consensus   chain.Consensus            // Consensus needs
	store       kvstore.KVStore            // State manager mock
	SolidState  state.VirtualState         // State manager mock
	StateOutput *ledgerstate.AliasOutput   // State manager mock
	Log         *logger.Logger
	mutex       sync.Mutex
}

func NewMockedEnv(t *testing.T, n, quorum uint16, debug bool) (*MockedEnv, *ledgerstate.Transaction) {
	return newMockedEnv(t, n, quorum, debug, false)
}

func NewMockedEnvWithMockedACS(t *testing.T, n, quorum uint16, debug bool) (*MockedEnv, *ledgerstate.Transaction) {
	return newMockedEnv(t, n, quorum, debug, true)
}

func newMockedEnv(t *testing.T, n, quorum uint16, debug, mockACS bool) (*MockedEnv, *ledgerstate.Transaction) {
	level := zapcore.InfoLevel
	if debug {
		level = zapcore.DebugLevel
	}
	log := testlogger.WithLevel(testlogger.NewLogger(t, "04:05.000"), level, false)
	var err error

	log.Infof("creating test environment with N = %d, T = %d", n, quorum)

	ret := &MockedEnv{
		T:      t,
		Quorum: quorum,
		Log:    log,
		Ledger: utxodb.New(),
		Nodes:  make([]*mockedNode, n),
	}
	if mockACS {
		ret.MockedACS = testchain.NewMockedACSRunner(quorum, log)
		log.Infof("running MOCKED ACS consensus")
	} else {
		log.Infof("running REAL ACS consensus")
	}

	ret.NetworkBehaviour = testutil.NewPeeringNetDynamic(log)

	log.Infof("running DKG and setting up mocked network..")
	nodeIDs, identities := testpeers.SetupKeys(n)
	ret.NodeIDs = nodeIDs
	ret.StateAddress, ret.DKSRegistries = testpeers.SetupDkgPregenerated(t, quorum, ret.NodeIDs, tcrypto.DefaultSuite())
	ret.NetworkProviders, ret.NetworkCloser = testpeers.SetupNet(ret.NodeIDs, identities, ret.NetworkBehaviour, log)

	ret.OriginatorKeyPair, ret.OriginatorAddress = ret.Ledger.NewKeyPairByIndex(0)
	_, err = ret.Ledger.RequestFunds(ret.OriginatorAddress)
	require.NoError(t, err)

	outputs := ret.Ledger.GetAddressOutputs(ret.OriginatorAddress)
	require.True(t, len(outputs) == 1)

	bals := colored.ToL1Map(colored.NewBalancesForIotas(100))

	txBuilder := utxoutil.NewBuilder(outputs...)
	err = txBuilder.AddNewAliasMint(bals, ret.StateAddress, state.OriginStateHash().Bytes())
	require.NoError(t, err)
	err = txBuilder.AddRemainderOutputIfNeeded(ret.OriginatorAddress, nil)
	require.NoError(t, err)
	originTx, err := txBuilder.BuildWithED25519(ret.OriginatorKeyPair)
	require.NoError(t, err)
	err = ret.Ledger.AddTransaction(originTx)
	require.NoError(t, err)

	ret.InitStateOutput, err = utxoutil.GetSingleChainedAliasOutput(originTx)
	require.NoError(t, err)

	ret.ChainID = *iscp.NewChainID(ret.InitStateOutput.GetAliasAddress())

	return ret, originTx
}

func (env *MockedEnv) CreateNodes(timers ConsensusTimers) {
	for i := range env.Nodes {
		env.Nodes[i] = env.NewNode(uint16(i), timers)
	}
}

func (env *MockedEnv) NewNode(nodeIndex uint16, timers ConsensusTimers) *mockedNode { //nolint:revive
	nodeID := env.NodeIDs[nodeIndex]
	log := env.Log.Named(nodeID)
	ret := &mockedNode{
		NodeID:    nodeID,
		Env:       env,
		NodeConn:  testchain.NewMockedNodeConnection("Node_" + nodeID),
		store:     mapdb.NewMapDB(),
		ChainCore: testchain.NewMockedChainCore(env.T, env.ChainID, log),
		stateSync: coreutil.NewChainStateSync(),
		Log:       log,
	}
	ret.ChainCore.OnGlobalStateSync(func() coreutil.ChainStateSync {
		return ret.stateSync
	})
	ret.ChainCore.OnGetStateReader(func() state.OptimisticStateReader {
		return state.NewOptimisticStateReader(ret.store, ret.stateSync)
	})
	ret.NodeConn.OnPostTransaction(func(tx *ledgerstate.Transaction) {
		env.mutex.Lock()
		defer env.mutex.Unlock()

		if _, already := env.Ledger.GetTransaction(tx.ID()); !already {
			if err := env.Ledger.AddTransaction(tx); err != nil {
				ret.Log.Error(err)
				return
			}
			stateOutput := transaction.GetAliasOutput(tx, env.ChainID.AsAddress())
			require.NotNil(env.T, stateOutput)

			ret.Log.Infof("stored transaction to the ledger: %s", tx.ID().Base58())
			for _, node := range env.Nodes {
				go func(n *mockedNode) {
					ret.mutex.Lock()
					defer ret.mutex.Unlock()
					n.StateOutput = stateOutput
					n.checkStateApproval()
				}(node)
			}
		} else {
			ret.Log.Infof("transaction already in the ledger: %s", tx.ID().Base58())
		}
	})
	ret.NodeConn.OnPullTransactionInclusionState(func(addr ledgerstate.Address, txid ledgerstate.TransactionID) {
		if _, already := env.Ledger.GetTransaction(txid); already {
			go ret.ChainCore.ReceiveMessage(&messages.InclusionStateMsg{
				TxID:  txid,
				State: ledgerstate.Confirmed,
			})
		}
	})
	mempoolMetrics := metrics.DefaultChainMetrics()
	ret.Mempool = mempool.New(ret.ChainCore.GetStateReader(), iscp.NewInMemoryBlobCache(), log, mempoolMetrics)

	cfg := &consensusTestConfigProvider{
		ownNetID:  nodeID,
		neighbors: env.NodeIDs,
	}
	//
	// Pass the ACS mock, if it was set in env.MockedACS.
	acs := make([]chain.AsynchronousCommonSubsetRunner, 0, 1)
	if env.MockedACS != nil {
		acs = append(acs, env.MockedACS)
	}
	cmtRec := &registry.CommitteeRecord{
		Address: env.StateAddress,
		Nodes:   env.NodeIDs,
	}
	cmt, err := committee.New(
		cmtRec,
		&env.ChainID,
		env.NetworkProviders[nodeIndex],
		cfg,
		env.DKSRegistries[nodeIndex],
		log,
		acs...,
	)
	require.NoError(env.T, err)
	cmt.Attach(ret.ChainCore)

	ret.StateOutput = env.InitStateOutput
	ret.SolidState, err = state.CreateOriginState(ret.store, &env.ChainID)
	ret.stateSync.SetSolidIndex(0)
	require.NoError(env.T, err)

	cons := New(ret.ChainCore, ret.Mempool, cmt, ret.NodeConn, true, timers)
	cons.vmRunner = testchain.NewMockedVMRunner(env.T, log)
	ret.Consensus = cons

	ret.ChainCore.OnReceiveAsynchronousCommonSubsetMsg(func(msg *messages.AsynchronousCommonSubsetMsg) {
		ret.Consensus.EventAsynchronousCommonSubsetMsg(msg)
	})
	ret.ChainCore.OnReceiveVMResultMsg(func(msg *messages.VMResultMsg) {
		ret.Consensus.EventVMResultMsg(msg)
	})
	ret.ChainCore.OnReceiveInclusionStateMsg(func(msg *messages.InclusionStateMsg) {
		ret.Consensus.EventInclusionsStateMsg(msg)
	})
	ret.ChainCore.OnReceiveStateCandidateMsg(func(msg *messages.StateCandidateMsg) {
		ret.mutex.Lock()
		defer ret.mutex.Unlock()
		newState := msg.State
		ret.Log.Infof("chainCore.StateCandidateMsg: state hash: %s, approving output: %s",
			msg.State.StateCommitment(), iscp.OID(msg.ApprovingOutputID))

		if ret.SolidState != nil && ret.SolidState.BlockIndex() == newState.BlockIndex() {
			ret.Log.Debugf("new state already committed for index %d", newState.BlockIndex())
			return
		}
		err := newState.Commit()
		require.NoError(env.T, err)

		ret.SolidState = newState
		ret.Log.Debugf("committed new state for index %d", newState.BlockIndex())

		ret.checkStateApproval()
	})
	ret.ChainCore.OnReceivePeerMessage(func(msg *peering.PeerMessage) {
		var err error
		if msg.MsgType == messages.MsgSignedResult {
			decoded := messages.SignedResultMsg{}
			if err = decoded.Read(bytes.NewReader(msg.MsgData)); err == nil {
				decoded.SenderIndex = msg.SenderIndex
				ret.Consensus.EventSignedResultMsg(&decoded)
			}
		}
		if err != nil {
			ret.Log.Errorf("unexpected peer message type = %d", msg.MsgType)
		}
	})
	return ret
}

func (env *MockedEnv) nodeCount() int {
	return len(env.NodeIDs)
}

func (env *MockedEnv) SetInitialConsensusState() {
	env.mutex.Lock()
	defer env.mutex.Unlock()

	for _, node := range env.Nodes {
		go func(n *mockedNode) {
			if n.SolidState != nil && n.SolidState.BlockIndex() == 0 {
				n.EventStateTransition()
			}
		}(node)
	}
}

func (n *mockedNode) checkStateApproval() {
	if n.SolidState == nil || n.StateOutput == nil {
		return
	}
	if n.SolidState.BlockIndex() != n.StateOutput.GetStateIndex() {
		return
	}
	stateHash, err := hashing.HashValueFromBytes(n.StateOutput.GetStateData())
	require.NoError(n.Env.T, err)
	require.EqualValues(n.Env.T, stateHash, n.SolidState.StateCommitment())

	reqIDsForLastState := make([]iscp.RequestID, 0)
	prefix := kv.Key(util.Uint32To4Bytes(n.SolidState.BlockIndex()))
	err = n.SolidState.KVStoreReader().Iterate(prefix, func(key kv.Key, value []byte) bool {
		reqid, err := iscp.RequestIDFromBytes(value)
		require.NoError(n.Env.T, err)
		reqIDsForLastState = append(reqIDsForLastState, reqid)
		return true
	})
	require.NoError(n.Env.T, err)
	n.Mempool.RemoveRequests(reqIDsForLastState...)

	n.Log.Infof("STATE APPROVED (%d reqs). Index: %d, State output: %s",
		len(reqIDsForLastState), n.SolidState.BlockIndex(), iscp.OID(n.StateOutput.ID()))

	n.EventStateTransition()
}

func (n *mockedNode) EventStateTransition() {
	n.Log.Debugf("EventStateTransition")

	n.ChainCore.GlobalStateSync().SetSolidIndex(n.SolidState.BlockIndex())

	n.Consensus.EventStateTransitionMsg(&messages.StateTransitionMsg{
		State:          n.SolidState.Clone(),
		StateOutput:    n.StateOutput,
		StateTimestamp: time.Now(),
	})
}

func (env *MockedEnv) StartTimers() {
	for _, n := range env.Nodes {
		n.StartTimer()
	}
}

func (n *mockedNode) StartTimer() {
	n.Log.Debugf("started timer..")
	go func() {
		counter := 0
		for {
			n.Consensus.EventTimerMsg(messages.TimerTick(counter))
			counter++
			time.Sleep(50 * time.Millisecond)
		}
	}()
}

func (env *MockedEnv) WaitTimerTick(until int) error {
	checkTimerTickFun := func(node *mockedNode) bool {
		snap := node.Consensus.GetStatusSnapshot()
		if snap != nil && snap.TimerTick >= until {
			return true
		}
		return false
	}
	return env.WaitForEventFromNodes("TimerTick", checkTimerTickFun)
}

func (env *MockedEnv) WaitStateIndex(quorum int, stateIndex uint32, timeout ...time.Duration) error {
	checkStateIndexFun := func(node *mockedNode) bool {
		snap := node.Consensus.GetStatusSnapshot()
		if snap != nil && snap.StateIndex >= stateIndex {
			return true
		}
		return false
	}
	return env.WaitForEventFromNodesQuorum("stateIndex", quorum, checkStateIndexFun, timeout...)
}

func (env *MockedEnv) WaitMempool(numRequests int, quorum int, timeout ...time.Duration) error { //nolint:gocritic
	checkMempoolFun := func(node *mockedNode) bool {
		snap := node.Consensus.GetStatusSnapshot()
		if snap != nil && snap.Mempool.InPoolCounter >= numRequests && snap.Mempool.OutPoolCounter >= numRequests {
			return true
		}
		return false
	}
	return env.WaitForEventFromNodesQuorum("mempool", quorum, checkMempoolFun, timeout...)
}

func (env *MockedEnv) WaitForEventFromNodes(waitName string, nodeConditionFun func(node *mockedNode) bool, timeout ...time.Duration) error {
	return env.WaitForEventFromNodesQuorum(waitName, env.nodeCount(), nodeConditionFun, timeout...)
}

func (env *MockedEnv) WaitForEventFromNodesQuorum(waitName string, quorum int, isEventOccuredFun func(node *mockedNode) bool, timeout ...time.Duration) error {
	to := 10 * time.Second
	if len(timeout) > 0 {
		to = timeout[0]
	}
	ch := make(chan int)
	nodeCount := env.nodeCount()
	deadline := time.Now().Add(to)
	for _, n := range env.Nodes {
		go func(node *mockedNode) {
			for time.Now().Before(deadline) {
				if isEventOccuredFun(node) {
					ch <- 1
				}
				time.Sleep(10 * time.Millisecond)
			}
			ch <- 0
		}(n)
	}
	var sum, total int
	for n := range ch {
		sum += n
		total++
		if sum >= quorum {
			return nil
		}
		if total >= nodeCount {
			return fmt.Errorf("Wait for %s: test timed out", waitName)
		}
	}
	return fmt.Errorf("WaitMempool: timeout expired %v", to)
}

func (env *MockedEnv) PostDummyRequests(n int, randomize ...bool) {
	reqs := make([]iscp.Request, n)
	for i := 0; i < n; i++ {
		reqs[i] = solo.NewCallParams("dummy", "dummy", "c", i).
			NewRequestOffLedger(env.OriginatorKeyPair)
	}
	rnd := len(randomize) > 0 && randomize[0]
	for _, n := range env.Nodes {
		if rnd {
			for _, req := range reqs {
				go func(node *mockedNode, r iscp.Request) {
					time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
					node.Mempool.ReceiveRequests(r)
				}(n, req)
			}
		} else {
			n.Mempool.ReceiveRequests(reqs...)
		}
	}
}

// TODO: should this object be obtained from peering.NetworkProvider?
// Or should registry.PeerNetworkConfigProvider methods methods be part of
// peering.NetworkProvider interface
type consensusTestConfigProvider struct {
	ownNetID  string
	neighbors []string
}

func (p *consensusTestConfigProvider) OwnNetID() string {
	return p.ownNetID
}

func (p *consensusTestConfigProvider) PeeringPort() int {
	return 0 // Anything
}

func (p *consensusTestConfigProvider) Neighbors() []string {
	return p.neighbors
}

func (p *consensusTestConfigProvider) String() string {
	return fmt.Sprintf("consensusTestConfigProvider( ownNetID: %s, neighbors: %+v )", p.OwnNetID(), p.Neighbors())
}
