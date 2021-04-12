package statemgr

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
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

	require.Nil(t, node0.StateManager.(*stateManager).solidState)
	require.EqualValues(t, 1, len(node0.StateManager.(*stateManager).blockCandidates))
	require.EqualValues(t, 0, env.Peers.NumPeers())
	env.AddNode(node0)
	require.EqualValues(t, 1, env.Peers.NumPeers())

	node0.StartTimer()
	node0.WaitSyncBlockIndex(0, 1*time.Second)

	require.Panics(t, func() {
		env.AddNode(node0)
	})

	node1 := env.NewMockedNode(1)
	require.NotPanics(t, func() {
		env.AddNode(node1)
	})
	require.EqualValues(t, 2, env.Peers.NumPeers())
	node1.StateManager.Ready().MustWait()

	require.Nil(t, node1.StateManager.(*stateManager).solidState)
	require.EqualValues(t, 1, len(node1.StateManager.(*stateManager).blockCandidates))

	node1.StartTimer()
	node1.WaitSyncBlockIndex(0, 1*time.Second)
	//time.Sleep(200 * time.Millisecond)

	env.RemoveNode(0)
	require.EqualValues(t, 1, env.Peers.NumPeers())

	env.AddNode(node0)
	require.EqualValues(t, 2, env.Peers.NumPeers())
}

func TestGetInitialState(t *testing.T) {
	env, originTx := NewMockedEnv(t, false)
	node := env.NewMockedNode(0)
	node.StateManager.Ready().MustWait()
	require.Nil(t, node.StateManager.(*stateManager).solidState)
	require.EqualValues(t, 1, len(node.StateManager.(*stateManager).blockCandidates))

	node.StartTimer()

	originOut, err := utxoutil.GetSingleChainedAliasOutput(originTx)
	require.NoError(t, err)

	env.AddNode(node)
	manager := node.StateManager.(*stateManager)

	time.Sleep(200 * time.Millisecond)
	require.True(t, originOut.Compare(manager.stateOutput) == 0)
	require.True(t, manager.stateOutput.GetStateIndex() == 0)
	require.EqualValues(t, manager.solidState.Hash(), state.OriginStateHash())
}

func TestGetNextState(t *testing.T) {
	env, originTx := NewMockedEnv(t, false)
	node := env.NewMockedNode(0)
	node.StateManager.Ready().MustWait()
	require.Nil(t, node.StateManager.(*stateManager).solidState)
	require.True(t, len(node.StateManager.(*stateManager).blockCandidates) == 1)

	node.StartTimer()

	originOut, err := utxoutil.GetSingleChainedAliasOutput(originTx)
	require.NoError(t, err)

	env.AddNode(node)
	manager := node.StateManager.(*stateManager)

	time.Sleep(200 * time.Millisecond)
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
	time.Sleep(200 * time.Millisecond)

	require.EqualValues(t, 1, manager.stateOutput.GetStateIndex())
	require.EqualValues(t, manager.solidState.Hash().Bytes(), manager.stateOutput.GetStateData())
	require.EqualValues(t, 0, len(manager.blockCandidates))
}
