package consensus

import (
	"bytes"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/chain/mempool"
	"github.com/iotaledger/wasp/packages/registry"

	"github.com/iotaledger/wasp/packages/util"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/transaction"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/solo"

	"github.com/iotaledger/wasp/packages/testutil"

	"go.dedis.ch/kyber/v3"

	"github.com/iotaledger/wasp/packages/testutil/testpeers"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxodb"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/committee"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3/pairing"
	"go.uber.org/zap/zapcore"
)

type mockedEnv struct {
	Suite             *pairing.SuiteBn256
	T                 *testing.T
	N                 uint16
	Quorum            uint16
	Neighbors         []string
	Log               *logger.Logger
	Ledger            *utxodb.UtxoDB
	OriginatorKeyPair *ed25519.KeyPair
	OriginatorAddress ledgerstate.Address
	StateAddress      ledgerstate.Address
	PubKeys           []kyber.Point
	PrivKeys          []kyber.Scalar
	NetworkProviders  []peering.NetworkProvider
	DKSRegistries     []registry.DKShareRegistryProvider
	store             kvstore.KVStore
	SolidState        state.VirtualState
	StateReader       state.StateReader
	StateOutput       *ledgerstate.AliasOutput
	RequestIDsLast    []coretypes.RequestID
	NodeConn          []*testchain.MockedNodeConn
	MockedACS         chain.AsynchronousCommonSubsetRunner
	ChainID           coretypes.ChainID
	mutex             sync.Mutex
	Nodes             []*mockedNode
	push              bool
}

type mockedNode struct {
	OwnIndex  uint16
	Env       *mockedEnv
	ChainCore *testchain.MockedChainCore
	Mempool   chain.Mempool
	Consensus *consensus
	Log       *logger.Logger
}

func NewMockedEnv(t *testing.T, n, quorum uint16, debug bool) (*mockedEnv, *ledgerstate.Transaction) {
	return newMockedEnv(t, n, quorum, debug, false)
}

func NewMockedEnvWithMockedACS(t *testing.T, n, quorum uint16, debug bool) (*mockedEnv, *ledgerstate.Transaction) {
	return newMockedEnv(t, n, quorum, debug, true)
}

func newMockedEnv(t *testing.T, n, quorum uint16, debug bool, mockACS bool) (*mockedEnv, *ledgerstate.Transaction) {
	level := zapcore.InfoLevel
	if debug {
		level = zapcore.DebugLevel
	}
	log := testlogger.WithLevel(testlogger.NewLogger(t, "04:05.000"), level, false)
	var err error

	log.Infof("creating test environment with N = %d, T = %d", n, quorum)

	neighbors := make([]string, n)
	for i := range neighbors {
		neighbors[i] = fmt.Sprintf("localhost:%d", 4000+i)
	}

	ret := &mockedEnv{
		Suite:     pairing.NewSuiteBn256(),
		T:         t,
		N:         n,
		Quorum:    quorum,
		Neighbors: neighbors,
		Log:       log,
		Ledger:    utxodb.New(),
		NodeConn:  make([]*testchain.MockedNodeConn, n),
		Nodes:     make([]*mockedNode, n),
	}
	if mockACS {
		ret.MockedACS = testchain.NewMockedACSRunner(quorum, log)
		log.Infof("running MOCKED ACS consensus")
	} else {
		log.Infof("running REAL ACS consensus")
	}

	for i := range ret.NodeConn {
		func(j int) {
			nconn := testchain.NewMockedNodeConnection(fmt.Sprintf("nodecon-%d", j))
			ret.NodeConn[j] = nconn
			nconn.OnPostTransaction(func(tx *ledgerstate.Transaction) {
				ret.receiveNewTransaction(tx, uint16(j))
			})
			nconn.OnPullBacklog(func(addr *ledgerstate.AliasAddress) {
				// TODO
			})
			nconn.OnPullTransactionInclusionState(func(addr ledgerstate.Address, txid ledgerstate.TransactionID) {
				if _, already := ret.Ledger.GetTransaction(txid); already {
					go ret.Nodes[j].ChainCore.ReceiveMessage(&chain.InclusionStateMsg{
						TxID:  txid,
						State: ledgerstate.Confirmed,
					})
				}
			})
		}(i)
	}

	log.Infof("running DKG and setting up mocked network..")
	_, ret.PubKeys, ret.PrivKeys = testpeers.SetupKeys(n, ret.Suite)
	ret.StateAddress, ret.DKSRegistries = testpeers.SetupDkg(t, quorum, neighbors, ret.PubKeys, ret.PrivKeys, ret.Suite, log.Named("dkg"))
	ret.NetworkProviders = testpeers.SetupNet(neighbors, ret.PubKeys, ret.PrivKeys, testutil.NewPeeringNetReliable(), log)

	ret.OriginatorKeyPair, ret.OriginatorAddress = ret.Ledger.NewKeyPairByIndex(0)
	_, err = ret.Ledger.RequestFunds(ret.OriginatorAddress)
	require.NoError(t, err)

	outputs := ret.Ledger.GetAddressOutputs(ret.OriginatorAddress)
	require.True(t, len(outputs) == 1)

	bals := map[ledgerstate.Color]uint64{ledgerstate.ColorIOTA: 100}
	txBuilder := utxoutil.NewBuilder(outputs...)
	err = txBuilder.AddNewAliasMint(bals, ret.StateAddress, state.OriginStateHash().Bytes())
	require.NoError(t, err)
	err = txBuilder.AddRemainderOutputIfNeeded(ret.OriginatorAddress, nil)
	require.NoError(t, err)
	originTx, err := txBuilder.BuildWithED25519(ret.OriginatorKeyPair)
	require.NoError(t, err)
	err = ret.Ledger.AddTransaction(originTx)
	require.NoError(t, err)

	ret.StateOutput, err = utxoutil.GetSingleChainedAliasOutput(originTx)
	require.NoError(t, err)

	ret.ChainID = *coretypes.NewChainID(ret.StateOutput.GetAliasAddress())

	ret.store = mapdb.NewMapDB()
	ret.SolidState, err = state.CreateOriginState(ret.store, &ret.ChainID)
	require.NoError(t, err)
	ret.StateReader, err = state.NewStateReader(ret.store)
	require.NoError(t, err)

	for i := range ret.Nodes {
		ret.Nodes[i] = ret.newNode(uint16(i))
	}
	return ret, originTx
}

func (env *mockedEnv) newNode(i uint16) *mockedNode {
	log := env.Log.Named(fmt.Sprintf("%d", i))
	chainCore := testchain.NewMockedChainCore(env.ChainID, log)
	mpool := mempool.New(env.StateReader, coretypes.NewInMemoryBlobCache(), log)
	mockCommitteeRegistry := testchain.NewMockedCommitteeRegistry(env.Neighbors)
	cfg, err := peering.NewStaticPeerNetworkConfigProvider(env.Neighbors[i], 4000+int(i), env.Neighbors...)
	require.NoError(env.T, err)
	//
	// Pass the ACS mock, if it was set in env.MockedACS.
	acs := make([]chain.AsynchronousCommonSubsetRunner, 0)
	if env.MockedACS != nil {
		acs = append(acs, env.MockedACS)
	}
	committee, err := committee.New(
		env.StateAddress,
		&env.ChainID,
		env.NetworkProviders[i],
		cfg,
		env.DKSRegistries[i],
		mockCommitteeRegistry,
		log,
		acs...,
	)
	require.NoError(env.T, err)

	committee.Attach(chainCore)
	ret := &mockedNode{
		OwnIndex:  i,
		Env:       env,
		ChainCore: chainCore,
		Mempool:   mpool,
		Consensus: New(chainCore, mpool, committee, env.NodeConn[i], log),
		Log:       log,
	}

	ret.Consensus.vmRunner = testchain.NewMockedVMRunner(env.T, log)
	chainCore.OnReceiveMessage(func(msg interface{}) {
		switch msg := msg.(type) {
		case *chain.AsynchronousCommonSubsetMsg:
			ret.Consensus.EventAsynchronousCommonSubsetMsg(msg)
		case *chain.VMResultMsg:
			ret.Consensus.EventVMResultMsg(msg)
		case *chain.StateCandidateMsg:
			ret.Log.Infof("chainCore.StateCandidateMsg: state hash: %s, approving output: %s",
				msg.State.Hash(), coretypes.OID(msg.ApprovingOutputID))
			ret.Env.receiveStateCandidate(msg.State, i)
		case *chain.InclusionStateMsg:
			ret.Consensus.EventInclusionsStateMsg(msg)
		case *peering.PeerMessage:
			ret.processPeerMessage(msg)
		default:
			ret.Log.Errorf("chainCore: unexpected message type: %T", msg)
		}
	})
	return ret
}

func (env *mockedEnv) receiveStateCandidate(newState state.VirtualState, from uint16) {
	env.mutex.Lock()
	defer env.mutex.Unlock()

	if env.SolidState != nil && env.SolidState.BlockIndex() == newState.BlockIndex() {
		env.Log.Debugf("node #%d: new state already committed for index %d", from, newState.BlockIndex())
		return
	}
	err := newState.Commit()
	require.NoError(env.T, err)

	env.SolidState = newState
	env.Log.Debugf("node #%d: committed new state for index %d", from, newState.BlockIndex())

	env.checkStateApproval(from)
}

func (env *mockedEnv) receiveNewTransaction(tx *ledgerstate.Transaction, from uint16) {
	env.mutex.Lock()
	defer env.mutex.Unlock()

	if _, already := env.Ledger.GetTransaction(tx.ID()); !already {
		if err := env.Ledger.AddTransaction(tx); err != nil {
			env.Log.Error(err)
			return
		}
		env.StateOutput = transaction.GetAliasOutput(tx, env.ChainID.AsAddress())
		require.NotNil(env.T, env.StateOutput)

		env.Log.Infof("node #%d: stored transaction to the ledger: %s", from, tx.ID().Base58())
		env.checkStateApproval(from)
	} else {
		env.Log.Infof("node #%d: transaction already in the ledger: %s", from, tx.ID().Base58())
	}
}

func (env *mockedEnv) checkStateApproval(from uint16) {
	if env.SolidState == nil || env.StateOutput == nil {
		return
	}
	if env.SolidState.BlockIndex() != env.StateOutput.GetStateIndex() {
		return
	}
	stateHash, err := hashing.HashValueFromBytes(env.StateOutput.GetStateData())
	require.NoError(env.T, err)
	require.EqualValues(env.T, stateHash, env.SolidState.Hash())

	env.RequestIDsLast = env.getReqIDsForLastState()

	env.Log.Infof("STATE APPROVED (%d reqs). Index: %d, State output: %s (from node #%d)",
		len(env.RequestIDsLast), env.SolidState.BlockIndex(), coretypes.OID(env.StateOutput.ID()), from)

	env.eventStateTransition()
}

func (env *mockedNode) processPeerMessage(msg *peering.PeerMessage) {
	var err error
	switch msg.MsgType {
	case chain.MsgSignedResult:
		decoded := chain.SignedResultMsg{}
		if err = decoded.Read(bytes.NewReader(msg.MsgData)); err == nil {
			decoded.SenderIndex = msg.SenderIndex
			env.Consensus.EventSignedResultMsg(&decoded)
		}
	}
	if err != nil {
		env.Log.Errorf("unexpected peer message type = %d", msg.MsgType)
	}
}

func (env *mockedEnv) eventStateTransition() {
	env.Log.Debugf("eventStateTransition")
	nowis := time.Now()
	solidState := env.SolidState.Clone()
	stateOutput := env.StateOutput

	for _, node := range env.Nodes {
		go func(n *mockedNode) {
			n.Mempool.RemoveRequests(env.RequestIDsLast...)
			n.ChainCore.GlobalSolidIndex().Store(solidState.BlockIndex())

			n.Consensus.EventStateTransitionMsg(&chain.StateTransitionMsg{
				State:          solidState.Clone(),
				StateOutput:    stateOutput,
				StateTimestamp: nowis,
			})
		}(node)
	}
}

func (env *mockedEnv) StartTimers() {
	for _, n := range env.Nodes {
		n.StartTimer()
	}
}

func (n *mockedNode) StartTimer() {
	n.Log.Debugf("started timer..")
	go func() {
		counter := 0
		for {
			n.Consensus.EventTimerMsg(chain.TimerTick(counter))
			counter++
			time.Sleep(50 * time.Millisecond)
		}
	}()
}

func (n *mockedNode) WaitTimerTick(until int) {
	for {
		snap := n.Consensus.GetStatusSnapshot()
		if snap == nil {
			continue
		}
		if snap.TimerTick >= until {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (env *mockedEnv) WaitTimerTick(until int) {
	var wg sync.WaitGroup
	wg.Add(int(env.N))
	for _, n := range env.Nodes {
		go func() {
			n.WaitTimerTick(until)
			wg.Done()
		}()
	}
	wg.Wait()
	env.Log.Infof("target timer tick #%d has been reached", until)
}

func (n *mockedNode) WaitStateIndex(until uint32, timeout ...time.Duration) error {
	deadline := time.Now().Add(10 * time.Second)
	if len(timeout) > 0 {
		deadline = time.Now().Add(timeout[0])
	}
	for {
		snap := n.Consensus.GetStatusSnapshot()
		if snap == nil {
			continue
		}
		if snap.StateIndex >= until {
			//n.Log.Debugf("reached index %d", until)
			return nil
		}
		time.Sleep(10 * time.Millisecond)
		if time.Now().After(deadline) {
			return fmt.Errorf("node %d: WaitStateIndex timeout", n.OwnIndex)
		}
	}
}

func (n *mockedNode) WaitMempool(numRequests int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		snap := n.Consensus.GetStatusSnapshot()
		if snap == nil {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		if snap.Mempool.InCounter >= numRequests && snap.Mempool.OutCounter >= numRequests {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("node %d: WaitMempool timeout", n.OwnIndex)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (env *mockedEnv) WaitStateIndex(quorum int, stateIndex uint32, timeout ...time.Duration) error {
	ch := make(chan int)
	for _, n := range env.Nodes {
		go func(n *mockedNode) {
			if err := n.WaitStateIndex(stateIndex, timeout...); err != nil {
				ch <- 0
			} else {
				ch <- 1
			}
		}(n)
	}
	var sum int
	for n := range ch {
		sum += n
		if sum >= quorum {
			return nil
		}
	}
	return fmt.Errorf("WaitStateIndex: timeout")
}

func (env *mockedEnv) WaitMempool(numRequests int, quorum int, timeout ...time.Duration) error {
	to := 10 * time.Second
	if len(timeout) > 0 {
		to = timeout[0]
	}
	ch := make(chan int)
	for _, n := range env.Nodes {
		go func(node *mockedNode) {
			if err := node.WaitMempool(numRequests, to); err != nil {
				ch <- 0
			} else {
				ch <- 1
			}
		}(n)
	}
	var sum, total int
	for n := range ch {
		sum += n
		total++
		if sum >= quorum {
			return nil
		}
		if total >= len(env.Nodes) {
			break
		}
	}
	return fmt.Errorf("WaitMempool: timeout expired %v", to)
}

func (env *mockedEnv) getReqIDsForLastState() []coretypes.RequestID {
	ret := make([]coretypes.RequestID, 0)
	prefix := kv.Key(util.Uint32To4Bytes(env.SolidState.BlockIndex()))
	err := env.SolidState.KVStoreReader().Iterate(prefix, func(key kv.Key, value []byte) bool {
		reqid, err := coretypes.RequestIDFromBytes(value)
		require.NoError(env.T, err)
		ret = append(ret, reqid)
		return true
	})
	require.NoError(env.T, err)
	return ret
}

func (env *mockedEnv) postDummyRequests(n int, randomize ...bool) {
	reqs := make([]coretypes.Request, n)
	for i := 0; i < n; i++ {
		reqs[i] = solo.NewCallParams("dummy", "dummy", "c", i).
			NewRequestOffLedger(env.OriginatorKeyPair)
	}
	rnd := len(randomize) > 0 && randomize[0]
	for _, n := range env.Nodes {
		if rnd {
			for _, req := range reqs {
				go func(node *mockedNode, r coretypes.Request) {
					time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
					node.Mempool.ReceiveRequests(r)
				}(n, req)
			}
		} else {
			n.Mempool.ReceiveRequests(reqs...)
		}
	}
}
