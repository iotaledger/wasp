package statemgr

import (
	"bytes"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxodb"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3/pairing"
	"go.uber.org/zap/zapcore"
	"golang.org/x/xerrors"
)

//---------------------------------------------
//Tests if state manager is started and initialised correctly
func TestEnv(t *testing.T) {
	env, _ := NewMockedEnv(2, t, false)
	node0 := env.NewMockedNode(0, Timers{})
	node0.StateManager.Ready().MustWait()

	require.NotNil(t, node0.StateManager.(*stateManager).solidState)
	require.EqualValues(t, state.OriginStateHash(), node0.StateManager.(*stateManager).solidState.Hash())
	require.False(t, node0.StateManager.(*stateManager).syncingBlocks.hasBlockCandidates())
	env.AddNode(node0)

	node0.StartTimer()
	waitSyncBlockIndexAndCheck(1*time.Second, t, node0, 0)

	require.Panics(t, func() {
		env.AddNode(node0)
	})

	node1 := env.NewMockedNode(1, Timers{})
	require.NotPanics(t, func() {
		env.AddNode(node1)
	})
	node1.StateManager.Ready().MustWait()

	require.NotNil(t, node1.StateManager.(*stateManager).solidState)
	require.False(t, node1.StateManager.(*stateManager).syncingBlocks.hasBlockCandidates())
	require.EqualValues(t, state.OriginStateHash(), node1.StateManager.(*stateManager).solidState.Hash())

	node1.StartTimer()
	waitSyncBlockIndexAndCheck(1*time.Second, t, node1, 0)

	env.RemoveNode(node0)
	require.EqualValues(t, 1, len(env.Nodes))

	env.AddNode(node0)
	require.EqualValues(t, 2, len(env.Nodes))
}

func TestGetInitialState(t *testing.T) {
	env, originTx := NewMockedEnv(1, t, false)
	node := env.NewMockedNode(0, Timers{})
	node.StateManager.Ready().MustWait()
	require.NotNil(t, node.StateManager.(*stateManager).solidState)
	require.False(t, node.StateManager.(*stateManager).syncingBlocks.hasBlockCandidates())
	require.EqualValues(t, state.OriginStateHash(), node.StateManager.(*stateManager).solidState.Hash())

	node.StartTimer()

	originOut, err := utxoutil.GetSingleChainedAliasOutput(originTx)
	require.NoError(t, err)

	env.AddNode(node)
	manager := node.StateManager.(*stateManager)

	syncInfo := waitSyncBlockIndexAndCheck(3*time.Second, t, node, 0)
	require.True(t, originOut.Compare(manager.stateOutput) == 0)
	require.True(t, manager.stateOutput.GetStateIndex() == 0)
	require.EqualValues(t, manager.solidState.Hash(), state.OriginStateHash())
	require.EqualValues(t, 0, syncInfo.SyncedBlockIndex)
	require.EqualValues(t, 0, syncInfo.StateOutputBlockIndex)
}

func TestGetNextState(t *testing.T) {
	env, originTx := NewMockedEnv(1, t, false)
	node := env.NewMockedNode(0, Timers{}.SetPullStateNewBlockDelay(50*time.Millisecond))
	node.StateManager.Ready().MustWait()
	require.NotNil(t, node.StateManager.(*stateManager).solidState)
	require.False(t, node.StateManager.(*stateManager).syncingBlocks.hasBlockCandidates())
	require.EqualValues(t, state.OriginStateHash(), node.StateManager.(*stateManager).solidState.Hash())

	node.StartTimer()

	originOut, err := utxoutil.GetSingleChainedAliasOutput(originTx)
	require.NoError(t, err)

	env.AddNode(node)
	manager := node.StateManager.(*stateManager)

	waitSyncBlockIndexAndCheck(1*time.Second, t, node, 0)
	require.True(t, originOut.Compare(manager.stateOutput) == 0)
	require.True(t, manager.stateOutput.GetStateIndex() == 0)
	require.EqualValues(t, manager.solidState.Hash(), state.OriginStateHash())

	//-------------------------------------------------------------

	currentState := manager.solidState
	require.NotNil(t, currentState)
	currentStateOutput := manager.stateOutput
	require.NotNil(t, currentState)
	currh := currentState.Hash()
	require.EqualValues(t, currh[:], currentStateOutput.GetStateData())

	node.StateTransition.NextState(currentState, currentStateOutput, time.Now())
	waitSyncBlockIndexAndCheck(3*time.Second, t, node, 1)

	require.EqualValues(t, 1, manager.stateOutput.GetStateIndex())
	require.EqualValues(t, manager.solidState.Hash().Bytes(), manager.stateOutput.GetStateData())
	require.False(t, manager.syncingBlocks.hasBlockCandidates())
}

func TestManyStateTransitionsPush(t *testing.T) {
	testManyStateTransitions(t, true)
}

func TestManyStateTransitionsNoPush(t *testing.T) {
	testManyStateTransitions(t, false)
}

// optionally, mocked node connection pushes new transactions to state managers or not.
// If not, state manager has to retrieve it with pull
func testManyStateTransitions(t *testing.T, pushStateToNodes bool) {
	env, _ := NewMockedEnv(1, t, false)
	env.SetPushStateToNodesOption(pushStateToNodes)

	timers := Timers{}
	if !pushStateToNodes {
		timers = timers.SetPullStateNewBlockDelay(50 * time.Millisecond)
	}

	node := env.NewMockedNode(0, timers)
	node.StateManager.Ready().MustWait()
	node.StartTimer()

	env.AddNode(node)

	const targetBlockIndex = 30
	node.OnStateTransitionMakeNewStateTransition(targetBlockIndex)
	waitSyncBlockIndexAndCheck(20*time.Second, t, node, targetBlockIndex)
}

// optionally, mocked node connection pushes new transactions to state managers or not.
// If not, state manager has to retrieve it with pull
func TestManyStateTransitionsSeveralNodes(t *testing.T) {
	env, _ := NewMockedEnv(2, t, true)
	env.SetPushStateToNodesOption(true)

	node := env.NewMockedNode(0, Timers{})
	node.StateManager.Ready().MustWait()
	node.StartTimer()

	env.AddNode(node)

	const targetBlockIndex = 10
	node.OnStateTransitionMakeNewStateTransition(targetBlockIndex)
	waitSyncBlockIndexAndCheck(10*time.Second, t, node, targetBlockIndex)

	node1 := env.NewMockedNode(1, Timers{})
	node1.StateManager.Ready().MustWait()
	node1.StartTimer()
	env.AddNode(node1)

	waitSyncBlockIndexAndCheck(10*time.Second, t, node1, targetBlockIndex)
}

func TestManyStateTransitionsManyNodes(t *testing.T) {
	numberOfCatchingPeers := 10
	env, _ := NewMockedEnv(numberOfCatchingPeers+1, t, true)
	env.SetPushStateToNodesOption(true)

	node := env.NewMockedNode(0, Timers{})
	node.StateManager.Ready().MustWait()
	node.StartTimer()

	env.AddNode(node)

	const targetBlockIndex = 5
	node.OnStateTransitionMakeNewStateTransition(targetBlockIndex)
	waitSyncBlockIndexAndCheck(10*time.Second, t, node, targetBlockIndex)

	catchingNodes := make([]*MockedNode, numberOfCatchingPeers)
	for i := 0; i < numberOfCatchingPeers; i++ {
		catchingNodes[i] = env.NewMockedNode(i+1, Timers{}.SetGetBlockRetry(200*time.Millisecond))
		catchingNodes[i].StateManager.Ready().MustWait()
	}
	for i := 0; i < numberOfCatchingPeers; i++ {
		catchingNodes[i].StartTimer()
	}
	for i := 0; i < numberOfCatchingPeers; i++ {
		env.AddNode(catchingNodes[i])
	}
	for i := 0; i < numberOfCatchingPeers; i++ {
		waitSyncBlockIndexAndCheck(20*time.Second, t, catchingNodes[i], targetBlockIndex)
	}
}

// Call to MsgGetConfirmetOutput does not return anything. Synchronisation must
// be done using stateOutput only.
func TestCatchUpNoConfirmedOutput(t *testing.T) {
	env, _ := NewMockedEnv(2, t, true)
	env.SetPushStateToNodesOption(true)

	node := env.NewMockedNode(0, Timers{})
	node.StateManager.Ready().MustWait()
	node.StartTimer()

	env.AddNode(node)

	const targetBlockIndex = 10
	node.OnStateTransitionMakeNewStateTransition(targetBlockIndex)
	node.NodeConn.OnPullConfirmedOutput(func(addr ledgerstate.Address, outputID ledgerstate.OutputID) {
	})
	waitSyncBlockIndexAndCheck(10*time.Second, t, node, targetBlockIndex)

	node1 := env.NewMockedNode(1, Timers{})
	node1.StateManager.Ready().MustWait()
	node1.StartTimer()
	env.AddNode(node1)

	waitSyncBlockIndexAndCheck(10*time.Second, t, node1, targetBlockIndex)
}

func TestNodeDisconnected(t *testing.T) {
	numberOfConnectedPeers := 5
	env, _ := NewMockedEnv(numberOfConnectedPeers+1, t, true)
	env.SetPushStateToNodesOption(false)

	createNodeFun := func(nodeIndex int) *MockedNode {
		result := env.NewMockedNode(nodeIndex, Timers{}.
			SetPullStateNewBlockDelay(150*time.Millisecond).
			SetPullStateRetry(150*time.Millisecond).
			SetGetBlockRetry(150*time.Millisecond),
		)
		result.StateManager.Ready().MustWait()
		result.StartTimer()
		env.AddNode(result)
		waitSyncBlockIndexAndCheck(10*time.Second, t, result, 0)
		return result
	}

	connectedNodes := make([]*MockedNode, numberOfConnectedPeers)
	for i := 0; i < numberOfConnectedPeers; i++ {
		connectedNodes[i] = createNodeFun(i)
	}
	disconnectedNode := createNodeFun(numberOfConnectedPeers)

	//Network is connected until state 3
	const targetBlockIndex1 = 3
	connectedNodes[0].OnStateTransitionMakeNewStateTransition(targetBlockIndex1)
	connectedNodes[0].MakeNewStateTransition()
	for i := 0; i < numberOfConnectedPeers; i++ {
		waitSyncBlockIndexAndCheck(10*time.Second, t, connectedNodes[i], targetBlockIndex1)
	}
	waitSyncBlockIndexAndCheck(10*time.Second, t, disconnectedNode, targetBlockIndex1)

	//Single node gets disconnected until state 6
	handlerName := "DisconnectedPeer"
	env.NetworkBehaviour.WithPeerDisconnected(&handlerName, disconnectedNode.NetID)
	const targetBlockIndex2 = 6
	connectedNodes[0].OnStateTransitionMakeNewStateTransition(targetBlockIndex2)
	connectedNodes[0].MakeNewStateTransition()
	for i := 0; i < numberOfConnectedPeers; i++ {
		waitSyncBlockIndexAndCheck(10*time.Second, t, connectedNodes[i], targetBlockIndex2)
	}

	//Network is reconnected until state 9, the node which was disconnected catches up
	env.NetworkBehaviour.RemoveHandler(handlerName)
	const targetBlockIndex3 = 9
	connectedNodes[0].OnStateTransitionMakeNewStateTransition(targetBlockIndex3)
	connectedNodes[0].MakeNewStateTransition()
	for i := 0; i < numberOfConnectedPeers; i++ {
		waitSyncBlockIndexAndCheck(10*time.Second, t, connectedNodes[i], targetBlockIndex3)
	}
	waitSyncBlockIndexAndCheck(10*time.Second, t, disconnectedNode, targetBlockIndex3)

	//Node, producing transitions, gets disconnected until state 12
	env.NetworkBehaviour.WithPeerDisconnected(&handlerName, disconnectedNode.NetID)
	const targetBlockIndex4 = 12
	connectedNodes[0].OnStateTransitionDoNothing()
	disconnectedNode.OnStateTransitionMakeNewStateTransition(targetBlockIndex4)
	disconnectedNode.MakeNewStateTransition()
	waitSyncBlockIndexAndCheck(10*time.Second, t, disconnectedNode, targetBlockIndex4)

	//Network is reconnected until state 15, other nodes catch up
	env.NetworkBehaviour.RemoveHandler(handlerName)
	const targetBlockIndex5 = 15
	disconnectedNode.OnStateTransitionMakeNewStateTransition(targetBlockIndex5)
	disconnectedNode.MakeNewStateTransition()
	for i := 0; i < numberOfConnectedPeers; i++ {
		waitSyncBlockIndexAndCheck(10*time.Second, t, connectedNodes[i], targetBlockIndex5)
	}
	waitSyncBlockIndexAndCheck(10*time.Second, t, disconnectedNode, targetBlockIndex5)
}

// 10 peers work in paralel. In every iteration random node is picked to produce
// a new state. Unreliable network is used, which delivers only 80% of messages,
// 25% o messages get delivered twice and messages are delayed up to 200 ms.
// Moreover, every 1-3s some random node gets disconnnected and later reconnected.
func TestCruelWorld(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	numberOfPeers := 10
	env, _ := NewMockedEnv(numberOfPeers, t, true)
	env.NetworkBehaviour.
		WithLosingChannel(nil, 80).
		WithRepeatingChannel(nil, 25).
		WithDelayingChannel(nil, 0*time.Millisecond, 200*time.Millisecond)
	env.SetPushStateToNodesOption(false)

	randFromIntervalFun := func(from int, till int) time.Duration {
		return time.Duration(from + rand.Intn(till-from))
	}
	nodes := make([]*MockedNode, numberOfPeers)
	for i := 0; i < numberOfPeers; i++ {
		nodes[i] = env.NewMockedNode(i, Timers{}.
			SetPullStateNewBlockDelay(randFromIntervalFun(200, 500)*time.Millisecond).
			SetPullStateRetry(randFromIntervalFun(50, 200)*time.Millisecond).
			SetGetBlockRetry(randFromIntervalFun(50, 200)*time.Millisecond),
		)
		nodes[i].StateManager.Ready().MustWait()
		nodes[i].StartTimer()
		env.AddNode(nodes[i])
	}

	var disconnectedNodes []string
	var mutex sync.Mutex
	go func() { //Connection cutter
		for {
			time.Sleep(randFromIntervalFun(1000, 3000) * time.Millisecond)
			mutex.Lock()
			nodeName := nodes[rand.Intn(numberOfPeers)].NetID
			env.NetworkBehaviour.WithPeerDisconnected(&nodeName, nodeName)
			env.Log.Debugf("Connection to node %v lost", nodeName)
			disconnectedNodes = append(disconnectedNodes, nodeName)
			mutex.Unlock()
		}
	}()

	go func() { //Connection restorer
		for {
			time.Sleep(randFromIntervalFun(500, 2000) * time.Millisecond)
			mutex.Lock()
			if len(disconnectedNodes) > 0 {
				env.NetworkBehaviour.RemoveHandler(disconnectedNodes[0])
				env.Log.Debugf("Connection to node %v restored", disconnectedNodes[0])
				disconnectedNodes[0] = ""
				disconnectedNodes = disconnectedNodes[1:]
			}
		}
	}()

	targetState := uint32(20)
	for i := uint32(0); i < targetState; i++ {
		randNode := nodes[rand.Intn(numberOfPeers)]
		waitSyncBlockIndexAndCheck(10*time.Second, t, randNode, i)
		randNode.MakeNewStateTransition()
	}

	for i := 0; i < numberOfPeers; i++ {
		waitSyncBlockIndexAndCheck(10*time.Second, t, nodes[i], targetState)
	}
}

func waitSyncBlockIndexAndCheck(duration time.Duration, t *testing.T, node *MockedNode, target uint32) *chain.SyncInfo {
	si, err := node.WaitSyncBlockIndex(target, duration)
	require.NoError(t, err)
	require.True(t, si.Synced)
	return si
}

type MockedEnv struct {
	T                 *testing.T
	Log               *logger.Logger
	Ledger            *utxodb.UtxoDB
	OriginatorKeyPair *ed25519.KeyPair
	OriginatorAddress ledgerstate.Address
	NodeIDs           []string
	NetworkProviders  []peering.NetworkProvider
	NetworkBehaviour  *testutil.PeeringNetDynamic
	ChainID           coretypes.ChainID
	mutex             sync.Mutex
	Nodes             map[string]*MockedNode
	push              bool
}

type MockedNode struct {
	NetID           string
	Env             *MockedEnv
	store           kvstore.KVStore
	NodeConn        *testchain.MockedNodeConn
	ChainCore       *testchain.MockedChainCore
	Peers           peering.PeerDomainProvider
	StateManager    chain.StateManager
	StateTransition *testchain.MockedStateTransition
	Log             *logger.Logger
}

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

	ret.NetworkBehaviour = testutil.NewPeeringNetDynamic(log)

	nodeIDs, pubKeys, privKeys := testpeers.SetupKeys(uint16(nodeCount), pairing.NewSuiteBn256())
	ret.NodeIDs = nodeIDs
	ret.NetworkProviders = testpeers.SetupNet(ret.NodeIDs, pubKeys, privKeys, ret.NetworkBehaviour, log)

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

func (env *MockedEnv) NewMockedNode(nodeIndex int, timers Timers) *MockedNode {
	nodeID := env.NodeIDs[nodeIndex]
	log := env.Log.Named(nodeID)
	peers, err := env.NetworkProviders[nodeIndex].PeerDomain(env.NodeIDs)
	require.NoError(env.T, err)
	ret := &MockedNode{
		NetID:     nodeID,
		Env:       env,
		NodeConn:  testchain.NewMockedNodeConnection("Node_" + nodeID),
		store:     mapdb.NewMapDB(),
		ChainCore: testchain.NewMockedChainCore(env.ChainID, log),
		Peers:     peers,
		Log:       log,
	}
	ret.StateManager = New(ret.store, ret.ChainCore, ret.Peers, ret.NodeConn, log, timers)
	ret.StateTransition = testchain.NewMockedStateTransition(env.T, env.OriginatorKeyPair)
	ret.StateTransition.OnNextState(func(vstate state.VirtualState, tx *ledgerstate.Transaction) {
		log.Debugf("MockedEnv.onNextState: state index %d", vstate.BlockIndex())
		go ret.StateManager.EventStateCandidateMsg(&chain.StateCandidateMsg{State: vstate})
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
	var peeringID peering.PeeringID = env.ChainID.Array()
	peers.Attach(&peeringID, func(recvEvent *peering.RecvEvent) {
		log.Debugf("MockedChain recvEvent from %v of type %v", recvEvent.From.NetID(), recvEvent.Msg.MsgType)
		rdr := bytes.NewReader(recvEvent.Msg.MsgData)

		switch recvEvent.Msg.MsgType {

		case chain.MsgGetBlock:
			msgt := &chain.GetBlockMsg{}
			if err := msgt.Read(rdr); err != nil {
				log.Error(err)
				return
			}

			msgt.SenderNetID = recvEvent.Msg.SenderNetID
			ret.StateManager.EventGetBlockMsg(msgt)

		case chain.MsgBlock:
			msgt := &chain.BlockMsg{}
			if err := msgt.Read(rdr); err != nil {
				log.Error(err)
				return
			}

			msgt.SenderNetID = recvEvent.Msg.SenderNetID
			ret.StateManager.EventBlockMsg(msgt)

		default:
			log.Errorf("MockedChain recvEvent: wrong msg type")
		}
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
		syncInfo = node.StateManager.GetStatusSnapshot()
		if syncInfo != nil && syncInfo.SyncedBlockIndex >= index {
			return syncInfo, nil
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (node *MockedNode) OnStateTransitionMakeNewStateTransition(limit uint32) {
	node.ChainCore.OnStateTransition(func(msg *chain.StateTransitionEventData) {
		chain.LogStateTransition(msg, node.Log)
		if msg.ChainOutput.GetStateIndex() < limit {
			go node.StateTransition.NextState(msg.VirtualState, msg.ChainOutput, time.Now())
		}
	})
}

func (node *MockedNode) OnStateTransitionDoNothing() {
	node.ChainCore.OnStateTransition(func(msg *chain.StateTransitionEventData) {})
}

func (node *MockedNode) MakeNewStateTransition() {
	node.StateTransition.NextState(node.StateManager.(*stateManager).solidState, node.StateManager.(*stateManager).stateOutput, time.Now())
}

func (env *MockedEnv) AddNode(node *MockedNode) {
	env.mutex.Lock()
	defer env.mutex.Unlock()

	if _, ok := env.Nodes[node.NetID]; ok {
		env.Log.Panicf("AddNode: duplicate node index %s", node.NetID)
	}
	env.Nodes[node.NetID] = node
}

func (env *MockedEnv) RemoveNode(node *MockedNode) {
	env.mutex.Lock()
	defer env.mutex.Unlock()
	delete(env.Nodes, node.NetID)
}
