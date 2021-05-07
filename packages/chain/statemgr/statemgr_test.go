package statemgr

import (
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/stretchr/testify/require"
)

//---------------------------------------------
//Tests if state manager is started and initialised correctly
func TestEnv(t *testing.T) {
	env, _ := NewMockedEnv(t, false)
	node0 := env.NewMockedNode("node0", nil)
	node0.SetupPeerGroupSimple()
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

	node1 := env.NewMockedNode("node1", nil)
	node1.SetupPeerGroupSimple()
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

	env.RemoveNode("node0")
	require.EqualValues(t, 1, len(env.Nodes))

	env.AddNode(node0)
	require.EqualValues(t, 2, len(env.Nodes))
}

func TestGetInitialState(t *testing.T) {
	env, originTx := NewMockedEnv(t, false)
	node := env.NewMockedNode("node0", nil)
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
	env, originTx := NewMockedEnv(t, false)
	node := env.NewMockedNode("node0", nil)
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
	env, _ := NewMockedEnv(t, false)
	env.SetPushStateToNodesOption(pushStateToNodes)

	node := env.NewMockedNode("node0", nil)
	node.StateManager.Ready().MustWait()
	node.StartTimer()

	env.AddNode(node)

	const targetBlockIndex = 30
	node.ChainCore.OnStateTransition(func(msg *chain.StateTransitionEventData) {
		chain.LogStateTransition(msg, node.Log)
		if msg.ChainOutput.GetStateIndex() < targetBlockIndex {
			go node.StateTransition.NextState(msg.VirtualState, msg.ChainOutput)
		}
	})
	si, err := node.WaitSyncBlockIndex(targetBlockIndex, 20*time.Second)
	require.NoError(t, err)
	require.True(t, si.Synced)
}

// optionally, mocked node connection pushes new transactions to state managers or not.
// If not, state manager has to retrieve it with pull
func TestManyStateTransitionsSeveralNodes(t *testing.T) {
	env, _ := NewMockedEnv(t, false)
	env.SetPushStateToNodesOption(true)

	allPeers := []string{"node0", "node1"}

	node := env.NewMockedNode("node0", allPeers)
	node.SetupPeerGroupSimple()
	node.StateManager.Ready().MustWait()
	node.StartTimer()

	env.AddNode(node)

	const targetBlockIndex = 10
	node.ChainCore.OnStateTransition(func(msg *chain.StateTransitionEventData) {
		chain.LogStateTransition(msg, node.Log)
		if msg.ChainOutput.GetStateIndex() < targetBlockIndex {
			go node.StateTransition.NextState(msg.VirtualState, msg.ChainOutput)
		}
	})
	env.NodeConn.OnPullConfirmedOutput(func(addr ledgerstate.Address, outputID ledgerstate.OutputID) {
		env.Log.Infof("MockedNodeConn.PullConfirmedOutput %v", coretypes.OID(outputID))
		env.mutex.Lock()
		defer env.mutex.Unlock()
		tx, foundTx := env.Ledger.GetTransaction(outputID.TransactionID())
		require.True(t, foundTx)
		outputIndex := outputID.OutputIndex()
		outputs := tx.Essence().Outputs()
		require.True(t, int(outputIndex) < len(outputs))
		output := outputs[outputIndex].UpdateMintingColor()
		require.NotNil(t, output)
		//TODO: avoid broadcast
		for _, node := range env.Nodes {
			go func(manager chain.StateManager, log *logger.Logger) {
				log.Infof("MockedNodeConn.PullConfirmedOutput: call EventOutputMsg")
				manager.EventOutputMsg(output)
			}(node.StateManager, node.Log)
		}
	})
	si, err := node.WaitSyncBlockIndex(targetBlockIndex, 10*time.Second)
	require.NoError(t, err)
	require.True(t, si.Synced)

	node1 := env.NewMockedNode("node1", allPeers)
	node1.SetupPeerGroupSimple()
	node1.StateManager.Ready().MustWait()
	node1.StartTimer()
	env.AddNode(node1)

	// node2 := env.NewMockedNode(2)
	// node2.StateManager.Ready().MustWait()
	// node2.StartTimer()
	// env.AddNode(node2)

	si, err = node1.WaitSyncBlockIndex(targetBlockIndex, 10*time.Second)
	require.NoError(t, err)
	require.True(t, si.Synced)

	// si, err = node2.WaitSyncBlockIndex(targetBlockIndex, 1*time.Second)
	// require.NoError(t, err)
	// require.True(t, si.Synced)
}
