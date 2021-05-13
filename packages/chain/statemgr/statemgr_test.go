package statemgr

import (
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/stretchr/testify/require"
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
	si, err := node0.WaitSyncBlockIndex(0, 1*time.Second)
	require.NoError(t, err)
	require.True(t, si.Synced)

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
	si, err = node1.WaitSyncBlockIndex(0, 1*time.Second)
	require.NoError(t, err)
	require.True(t, si.Synced)

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

	syncInfo, err := node.WaitSyncBlockIndex(0, 3*time.Second)
	require.NoError(t, err)
	require.True(t, syncInfo.Synced)
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

	si, err := node.WaitSyncBlockIndex(0, 1*time.Second)
	require.NoError(t, err)
	require.True(t, si.Synced)
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

	node.StateTransition.NextState(currentState, currentStateOutput)
	si, err = node.WaitSyncBlockIndex(1, 3*time.Second)
	require.NoError(t, err)
	require.True(t, si.Synced)

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
	si, err := node.WaitSyncBlockIndex(targetBlockIndex, 20*time.Second)
	require.NoError(t, err)
	require.True(t, si.Synced)
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
	si, err := node.WaitSyncBlockIndex(targetBlockIndex, 10*time.Second)
	require.NoError(t, err)
	require.True(t, si.Synced)

	node1 := env.NewMockedNode(1, Timers{})
	node1.StateManager.Ready().MustWait()
	node1.StartTimer()
	env.AddNode(node1)

	si, err = node1.WaitSyncBlockIndex(targetBlockIndex, 10*time.Second)
	require.NoError(t, err)
	require.True(t, si.Synced)
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
	si, err := node.WaitSyncBlockIndex(targetBlockIndex, 10*time.Second)
	require.NoError(t, err)
	require.True(t, si.Synced)

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
		si, err = catchingNodes[i].WaitSyncBlockIndex(targetBlockIndex, 10*time.Second)
		require.NoError(t, err)
		require.True(t, si.Synced)
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
	si, err := node.WaitSyncBlockIndex(targetBlockIndex, 10*time.Second)
	require.NoError(t, err)
	require.True(t, si.Synced)

	node1 := env.NewMockedNode(1, Timers{})
	node1.StateManager.Ready().MustWait()
	node1.StartTimer()
	env.AddNode(node1)

	si, err = node1.WaitSyncBlockIndex(targetBlockIndex, 10*time.Second)
	require.NoError(t, err)
	require.True(t, si.Synced)
}

func TestNodeDisconnected(t *testing.T) {
	numberOfConnectedPeers := 5
	env, _ := NewMockedEnv(numberOfConnectedPeers+1, t, true)
	env.SetPushStateToNodesOption(true)

	checkResultFun := func(node *MockedNode, target uint32) {
		si, err := node.WaitSyncBlockIndex(target, 10*time.Second)
		require.NoError(t, err)
		require.True(t, si.Synced)
	}
	createNodeFun := func(nodeIndex int) *MockedNode {
		result := env.NewMockedNode(nodeIndex, Timers{})
		result.StateManager.Ready().MustWait()
		result.StartTimer()
		env.AddNode(result)
		checkResultFun(result, 0)
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
		checkResultFun(connectedNodes[i], targetBlockIndex1)
	}
	checkResultFun(disconnectedNode, targetBlockIndex1)

	//Single node gets disconnected until state 6
	handlerName := "DisconnectedPeer"
	env.NetworkBehaviour.WithPeerDisconnected(&handlerName, disconnectedNode.NetID)
	const targetBlockIndex2 = 6
	connectedNodes[0].OnStateTransitionMakeNewStateTransition(targetBlockIndex2)
	connectedNodes[0].MakeNewStateTransition()
	for i := 0; i < numberOfConnectedPeers; i++ {
		checkResultFun(connectedNodes[i], targetBlockIndex2)
	}

	//Network is reconnected until state 9, the node which was disconnected catches up
	env.NetworkBehaviour.RemoveHandler(handlerName)
	const targetBlockIndex3 = 9
	connectedNodes[0].OnStateTransitionMakeNewStateTransition(targetBlockIndex3)
	connectedNodes[0].MakeNewStateTransition()
	for i := 0; i < numberOfConnectedPeers; i++ {
		checkResultFun(connectedNodes[i], targetBlockIndex3)
	}
	checkResultFun(disconnectedNode, targetBlockIndex3)

	//Node, producing transitions, gets disconnected until state 12
	env.NetworkBehaviour.WithPeerDisconnected(&handlerName, disconnectedNode.NetID)
	const targetBlockIndex4 = 12
	connectedNodes[0].OnStateTransitionDoNothing()
	disconnectedNode.OnStateTransitionMakeNewStateTransition(targetBlockIndex4)
	disconnectedNode.MakeNewStateTransition()
	checkResultFun(disconnectedNode, targetBlockIndex4)

	//Network is reconnected until state 15, other nodes catch up
	env.NetworkBehaviour.RemoveHandler(handlerName)
	const targetBlockIndex5 = 15
	disconnectedNode.OnStateTransitionMakeNewStateTransition(targetBlockIndex5)
	disconnectedNode.MakeNewStateTransition()
	for i := 0; i < numberOfConnectedPeers; i++ {
		checkResultFun(connectedNodes[i], targetBlockIndex5)
	}
	checkResultFun(disconnectedNode, targetBlockIndex5)
}
