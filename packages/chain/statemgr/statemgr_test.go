// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/stretchr/testify/require"
)

//---------------------------------------------
//Tests if state manager is started and initialized correctly
func TestEnv(t *testing.T) {
	env := NewMockedEnv(2, t, false)
	node0 := NewMockedNode(env, 0, NewStateManagerTimers())
	node0.StateManager.Ready().MustWait()

	require.NotNil(t, node0.StateManager.(*stateManager).solidState)
	require.True(t, state.EqualCommitments(state.OriginStateCommitment(), state.RootCommitment(node0.StateManager.(*stateManager).solidState.TrieNodeStore())))
	require.False(t, node0.StateManager.(*stateManager).syncingBlocks.hasBlockCandidates())
	env.AddNode(node0)

	node0.Start()
	waitSyncBlockIndexAndCheck(1*time.Second, t, node0, 0)

	require.Panics(t, func() {
		env.AddNode(node0)
	})

	node1 := NewMockedNode(env, 1, NewStateManagerTimers())
	require.NotPanics(t, func() {
		env.AddNode(node1)
	})
	node1.StateManager.Ready().MustWait()

	require.NotNil(t, node1.StateManager.(*stateManager).solidState)
	require.False(t, node1.StateManager.(*stateManager).syncingBlocks.hasBlockCandidates())
	require.True(t, state.EqualCommitments(state.OriginStateCommitment(), state.RootCommitment(node1.StateManager.(*stateManager).solidState.TrieNodeStore())))

	node1.Start()
	waitSyncBlockIndexAndCheck(1*time.Second, t, node1, 0)

	env.RemoveNode(node0)
	require.EqualValues(t, 1, len(env.Nodes))

	env.AddNode(node0)
	require.EqualValues(t, 2, len(env.Nodes))
}

func TestGetInitialState(t *testing.T) {
	env := NewMockedEnv(1, t, false)
	node := NewMockedNode(env, 0, NewStateManagerTimers())
	node.StateManager.Ready().MustWait()
	require.NotNil(t, node.StateManager.(*stateManager).solidState)
	require.False(t, node.StateManager.(*stateManager).syncingBlocks.hasBlockCandidates())
	require.True(t, state.EqualCommitments(state.OriginStateCommitment(), state.RootCommitment(node.StateManager.(*stateManager).solidState.TrieNodeStore())))
	env.AddNode(node)

	node.Start()
	syncInfo := waitSyncBlockIndexAndCheck(3*time.Second, t, node, 0)
	originOutput := env.Ledgers.GetLedger(env.ChainID).GetOriginOutput().GetAliasOutput()
	manager := node.StateManager.(*stateManager)
	require.True(t, isc.AliasOutputsEqual(originOutput, manager.stateOutput.GetAliasOutput()))
	require.True(t, manager.stateOutput.GetStateIndex() == 0)
	require.True(t, state.EqualCommitments(state.OriginStateCommitment(), state.RootCommitment(manager.solidState.TrieNodeStore())))
	require.EqualValues(t, 0, syncInfo.SyncedBlockIndex)
	require.EqualValues(t, 0, syncInfo.StateOutput.GetStateIndex())
}

func TestGetNextState(t *testing.T) {
	env := NewMockedEnv(1, t, false)
	originOutput := env.Ledgers.GetLedger(env.ChainID).GetOriginOutput().GetAliasOutput()
	timers := NewStateManagerTimers()
	timers.PullStateAfterStateCandidateDelay = 50 * time.Millisecond
	node := NewMockedNode(env, 0, timers)
	node.StateManager.Ready().MustWait()
	require.NotNil(t, node.StateManager.(*stateManager).solidState)
	require.False(t, node.StateManager.(*stateManager).syncingBlocks.hasBlockCandidates())
	require.True(t, state.EqualCommitments(state.OriginStateCommitment(), state.RootCommitment(node.StateManager.(*stateManager).solidState.TrieNodeStore())))
	env.AddNode(node)

	node.Start()
	waitSyncBlockIndexAndCheck(1*time.Second, t, node, 0)
	manager := node.StateManager.(*stateManager)
	require.True(t, isc.AliasOutputsEqual(originOutput, manager.stateOutput.GetAliasOutput()))
	require.True(t, manager.stateOutput.GetStateIndex() == 0)
	require.True(t, state.EqualCommitments(state.OriginStateCommitment(), state.RootCommitment(manager.solidState.TrieNodeStore())))

	//-------------------------------------------------------------

	currentState := manager.solidState
	require.NotNil(t, currentState)
	currentStateOutput := manager.stateOutput
	require.NotNil(t, currentState)
	currSH := state.RootCommitment(currentState.TrieNodeStore())
	currOH, err := state.L1CommitmentFromAliasOutput(currentStateOutput.GetAliasOutput())
	require.NoError(t, err)
	require.True(t, state.EqualCommitments(currSH, currOH.StateCommitment))

	node.NextState(currentState, currentStateOutput)
	waitSyncBlockIndexAndCheck(3*time.Second, t, node, 1)

	soc, err := state.L1CommitmentFromAliasOutput(manager.stateOutput.GetAliasOutput())
	require.NoError(t, err)
	require.EqualValues(t, 1, manager.stateOutput.GetStateIndex())
	require.True(t, state.EqualCommitments(state.RootCommitment(manager.solidState.TrieNodeStore()), soc.StateCommitment))
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
	env := NewMockedEnv(1, t, false)
	env.SetPushStateToNodesOption(pushStateToNodes)

	timers := NewStateManagerTimers()
	if !pushStateToNodes {
		timers.PullStateAfterStateCandidateDelay = 50 * time.Millisecond
	}

	node := NewMockedNode(env, 0, timers)
	node.StateManager.Ready().MustWait()
	env.AddNode(node)

	const targetBlockIndex = 30
	node.OnStateTransitionMakeNewStateTransition(targetBlockIndex)
	node.Start()
	waitSyncBlockIndexAndCheck(20*time.Second, t, node, targetBlockIndex)
}

// optionally, mocked node connection pushes new transactions to state managers or not.
// If not, state manager has to retrieve it with pull
func TestManyStateTransitionsSeveralNodes(t *testing.T) {
	env := NewMockedEnv(2, t, false)
	env.SetPushStateToNodesOption(true)

	node0 := NewMockedNode(env, 0, NewStateManagerTimers())
	node0.StateManager.Ready().MustWait()
	node0.StateManager.SetChainPeers([]*cryptolib.PublicKey{node0.PubKey})
	env.AddNode(node0)
	env.Log.Infof("TestManyStateTransitionsSeveralNodes: node0.PubKey=%v", node0.PubKey.String())

	const targetBlockIndex = 10
	node0.OnStateTransitionMakeNewStateTransition(targetBlockIndex)
	node0.Start()
	waitSyncBlockIndexAndCheck(10*time.Second, t, node0, targetBlockIndex)

	node1 := NewMockedNode(env, 1, NewStateManagerTimers())
	node1.StateManager.Ready().MustWait()
	node1.StateManager.SetChainPeers([]*cryptolib.PublicKey{node0.PubKey, node1.PubKey})
	node0.StateManager.SetChainPeers([]*cryptolib.PublicKey{node0.PubKey, node1.PubKey})
	env.AddNode(node1)
	env.Log.Infof("TestManyStateTransitionsSeveralNodes: node1.PubKey=%v", node1.PubKey.String())

	node1.Start()
	waitSyncBlockIndexAndCheck(10*time.Second, t, node1, targetBlockIndex)
}

func TestManyStateTransitionsManyNodes(t *testing.T) {
	numberOfCatchingPeers := 10
	env := NewMockedEnv(numberOfCatchingPeers+1, t, false)
	env.SetPushStateToNodesOption(true)

	allPubKeys := make([]*cryptolib.PublicKey, 0)

	node0 := NewMockedNode(env, 0, NewStateManagerTimers())
	node0.StateManager.Ready().MustWait()
	allPubKeys = append(allPubKeys, node0.PubKey)

	env.AddNode(node0)
	node0.StateManager.SetChainPeers(allPubKeys)

	const targetBlockIndex = 5
	node0.OnStateTransitionMakeNewStateTransition(targetBlockIndex)
	node0.Start()
	waitSyncBlockIndexAndCheck(10*time.Second, t, node0, targetBlockIndex)

	catchingNodes := make([]*MockedNode, numberOfCatchingPeers)
	for i := 0; i < numberOfCatchingPeers; i++ {
		timers := NewStateManagerTimers()
		timers.GetBlockRetry = 200 * time.Millisecond
		catchingNodes[i] = NewMockedNode(env, i+1, timers)
		catchingNodes[i].StateManager.Ready().MustWait()
		allPubKeys = append(allPubKeys, catchingNodes[i].PubKey)
	}
	node0.StateManager.SetChainPeers(allPubKeys)
	for i := 0; i < numberOfCatchingPeers; i++ {
		catchingNodes[i].StateManager.SetChainPeers(allPubKeys)
	}
	for i := 0; i < numberOfCatchingPeers; i++ {
		env.AddNode(catchingNodes[i])
		catchingNodes[i].Start()
	}
	for i := 0; i < numberOfCatchingPeers; i++ {
		waitSyncBlockIndexAndCheck(30*time.Second, t, catchingNodes[i], targetBlockIndex)
	}
}

// Call to MsgGetConfirmetOutput does not return anything. Synchronization must
// be done using stateOutput only.
func TestCatchUpNoConfirmedOutput(t *testing.T) {
	env := NewMockedEnv(2, t, false)
	env.SetPushStateToNodesOption(true)

	node0 := NewMockedNode(env, 0, NewStateManagerTimers())
	node0.StateManager.Ready().MustWait()
	node0.StateManager.SetChainPeers([]*cryptolib.PublicKey{node0.PubKey})
	env.AddNode(node0)

	const targetBlockIndex = 10
	node0.OnStateTransitionMakeNewStateTransition(targetBlockIndex)
	node0.NodeConn.SetPullOutputByIDAllowed(false)
	node0.Start()
	waitSyncBlockIndexAndCheck(10*time.Second, t, node0, targetBlockIndex)

	node1 := NewMockedNode(env, 1, NewStateManagerTimers())
	node1.StateManager.Ready().MustWait()
	node1.StateManager.SetChainPeers([]*cryptolib.PublicKey{node0.PubKey, node1.PubKey})
	node0.StateManager.SetChainPeers([]*cryptolib.PublicKey{node0.PubKey, node1.PubKey})
	env.AddNode(node1)

	node1.Start()
	waitSyncBlockIndexAndCheck(10*time.Second, t, node1, targetBlockIndex)
}

func TestNodeDisconnected(t *testing.T) {
	numberOfConnectedPeers := 5
	env := NewMockedEnv(numberOfConnectedPeers+1, t, false)
	env.SetPushStateToNodesOption(false)

	createNodeFun := func(nodeIndex int) *MockedNode {
		timers := NewStateManagerTimers()
		timers.PullStateAfterStateCandidateDelay = 150 * time.Millisecond
		timers.PullStateRetry = 150 * time.Millisecond
		timers.GetBlockRetry = 150 * time.Millisecond
		result := NewMockedNode(env, nodeIndex, timers)
		result.StateManager.Ready().MustWait()
		env.AddNode(result)
		result.Start()
		waitSyncBlockIndexAndCheck(10*time.Second, t, result, 0)
		return result
	}

	connectedNodes := make([]*MockedNode, numberOfConnectedPeers)
	for i := 0; i < numberOfConnectedPeers; i++ {
		connectedNodes[i] = createNodeFun(i)
	}
	disconnectedNode := createNodeFun(numberOfConnectedPeers)

	// Network is connected until state 3
	const targetBlockIndex1 = 3
	connectedNodes[0].OnStateTransitionMakeNewStateTransition(targetBlockIndex1)
	connectedNodes[0].MakeNewStateTransition()
	for i := 0; i < numberOfConnectedPeers; i++ {
		waitSyncBlockIndexAndCheck(10*time.Second, t, connectedNodes[i], targetBlockIndex1)
	}
	waitSyncBlockIndexAndCheck(10*time.Second, t, disconnectedNode, targetBlockIndex1)

	// Single node gets disconnected until state 6
	handlerName := "DisconnectedPeer"
	env.NetworkBehaviour.WithPeerDisconnected(&handlerName, disconnectedNode.PubKey)
	const targetBlockIndex2 = 6
	connectedNodes[0].OnStateTransitionMakeNewStateTransition(targetBlockIndex2)
	connectedNodes[0].MakeNewStateTransition()
	for i := 0; i < numberOfConnectedPeers; i++ {
		waitSyncBlockIndexAndCheck(10*time.Second, t, connectedNodes[i], targetBlockIndex2)
	}

	// Network is reconnected until state 9, the node which was disconnected catches up
	env.NetworkBehaviour.RemoveHandler(handlerName)
	const targetBlockIndex3 = 9
	connectedNodes[0].OnStateTransitionMakeNewStateTransition(targetBlockIndex3)
	connectedNodes[0].MakeNewStateTransition()
	for i := 0; i < numberOfConnectedPeers; i++ {
		waitSyncBlockIndexAndCheck(10*time.Second, t, connectedNodes[i], targetBlockIndex3)
	}
	waitSyncBlockIndexAndCheck(10*time.Second, t, disconnectedNode, targetBlockIndex3)

	// Node, producing transitions, gets disconnected until state 12
	env.NetworkBehaviour.WithPeerDisconnected(&handlerName, disconnectedNode.PubKey)
	const targetBlockIndex4 = 12
	connectedNodes[0].OnStateTransitionDoNothing()
	disconnectedNode.OnStateTransitionMakeNewStateTransition(targetBlockIndex4)
	disconnectedNode.MakeNewStateTransition()
	waitSyncBlockIndexAndCheck(10*time.Second, t, disconnectedNode, targetBlockIndex4)

	// Network is reconnected until state 15, other nodes catch up
	env.NetworkBehaviour.RemoveHandler(handlerName)
	const targetBlockIndex5 = 15
	disconnectedNode.OnStateTransitionMakeNewStateTransition(targetBlockIndex5)
	disconnectedNode.MakeNewStateTransition()
	for i := 0; i < numberOfConnectedPeers; i++ {
		waitSyncBlockIndexAndCheck(10*time.Second, t, connectedNodes[i], targetBlockIndex5)
	}
	waitSyncBlockIndexAndCheck(10*time.Second, t, disconnectedNode, targetBlockIndex5)
}

// 10 peers work in parallel. In every iteration random node is picked to produce
// a new state. Unreliable network is used, which delivers only 80% of messages,
// 25% o messages get delivered twice and messages are delayed up to 200 ms.
// Moreover, every 1-3s some random node gets disconnnected and later reconnected.
func TestCruelWorld(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	numberOfPeers := 10
	env := NewMockedEnv(numberOfPeers, t, false)
	env.NetworkBehaviour.
		WithLosingChannel(nil, 80).
		WithRepeatingChannel(nil, 25).
		WithDelayingChannel(nil, 0*time.Millisecond, 200*time.Millisecond)
	env.SetPushStateToNodesOption(false)

	rand.Seed(time.Now().UnixNano())
	randFromIntervalFun := func(from int, till int) time.Duration {
		return time.Duration(from + rand.Intn(till-from))
	}
	nodes := make([]*MockedNode, numberOfPeers)
	for i := 0; i < numberOfPeers; i++ {
		timers := NewStateManagerTimers()
		timers.PullStateAfterStateCandidateDelay = randFromIntervalFun(200, 500) * time.Millisecond
		timers.PullStateRetry = randFromIntervalFun(50, 200) * time.Millisecond
		timers.GetBlockRetry = randFromIntervalFun(50, 200) * time.Millisecond
		nodes[i] = NewMockedNode(env, i, timers)
		nodes[i].StateManager.Ready().MustWait()
		nodes[i].StateManager.SetChainPeers(env.NodePubKeys)
		env.AddNode(nodes[i])
		nodes[i].Start()
	}

	var disconnectedNodes []*cryptolib.PublicKey
	var mutex sync.Mutex
	go func() { // Connection cutter
		for {
			time.Sleep(randFromIntervalFun(1000, 3000) * time.Millisecond)
			mutex.Lock()
			nodePubkey := nodes[rand.Intn(numberOfPeers)].PubKey
			handlerID := nodePubkey.String()
			env.NetworkBehaviour.WithPeerDisconnected(&handlerID, nodePubkey)
			env.Log.Debugf("Connection to node %v lost", nodePubkey.String())
			disconnectedNodes = append(disconnectedNodes, nodePubkey)
			mutex.Unlock()
		}
	}()

	go func() { // Connection restorer
		for {
			time.Sleep(randFromIntervalFun(500, 2000) * time.Millisecond)
			mutex.Lock()
			if len(disconnectedNodes) > 0 {
				env.NetworkBehaviour.RemoveHandler(disconnectedNodes[0].String())
				env.Log.Debugf("Connection to node %v restored", disconnectedNodes[0])
				disconnectedNodes[0] = nil
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
