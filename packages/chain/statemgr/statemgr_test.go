package statemgr

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/state"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

//---------------------------------------------
//Tests if state manager is started and initialised correctly
func TestEnv(t *testing.T) {
	env, _ := NewMockedEnv(t, false)
	env.SetupPeerGroupSimple()
	node0 := env.NewMockedNode(0)
	node0.StateManager.Ready().MustWait()

	require.NotNil(t, node0.StateManager.(*stateManager).solidState)
	require.EqualValues(t, state.OriginStateHash(), node0.StateManager.(*stateManager).solidState.Hash())
	require.EqualValues(t, 0, len(node0.StateManager.(*stateManager).stateCandidates))
	require.EqualValues(t, 0, env.Peers.NumPeers())
	env.AddNode(node0)
	require.EqualValues(t, 1, env.Peers.NumPeers())

	node0.StartTimer()
	si, err := node0.WaitSyncBlockIndex(0, 1*time.Second)
	require.NoError(t, err)
	require.True(t, si.Synced)

	require.Panics(t, func() {
		env.AddNode(node0)
	})

	node1 := env.NewMockedNode(1)
	require.NotPanics(t, func() {
		env.AddNode(node1)
	})
	require.EqualValues(t, 2, env.Peers.NumPeers())
	node1.StateManager.Ready().MustWait()

	require.NotNil(t, node1.StateManager.(*stateManager).solidState)
	require.EqualValues(t, 0, len(node1.StateManager.(*stateManager).stateCandidates))
	require.EqualValues(t, state.OriginStateHash(), node1.StateManager.(*stateManager).solidState.Hash())

	node1.StartTimer()
	si, err = node1.WaitSyncBlockIndex(0, 1*time.Second)
	require.NoError(t, err)
	require.True(t, si.Synced)

	env.RemoveNode(0)
	require.EqualValues(t, 1, env.Peers.NumPeers())

	env.AddNode(node0)
	require.EqualValues(t, 2, env.Peers.NumPeers())
}

func TestGetInitialState(t *testing.T) {
	env, originTx := NewMockedEnv(t, false)
	node := env.NewMockedNode(0)
	node.StateManager.Ready().MustWait()
	require.NotNil(t, node.StateManager.(*stateManager).solidState)
	require.EqualValues(t, 0, len(node.StateManager.(*stateManager).stateCandidates))
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
	node := env.NewMockedNode(0)
	node.StateManager.Ready().MustWait()
	require.NotNil(t, node.StateManager.(*stateManager).solidState)
	require.EqualValues(t, 0, len(node.StateManager.(*stateManager).stateCandidates))
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
	require.EqualValues(t, 0, len(manager.stateCandidates))
}

// optionally, mocked node connection pushes new transactions to state managers or not.
// If not, state manager hash to retrieve it with pull
const pushStateToNodes = true

func TestManyStateTransitions(t *testing.T) {
	env, _ := NewMockedEnv(t, false)
	env.SetPushStateToNodesOption(pushStateToNodes)

	node := env.NewMockedNode(0)
	node.StateManager.Ready().MustWait()
	node.StartTimer()

	env.AddNode(node)

	const targetBlockIndex = 100
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
