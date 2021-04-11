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
func TestStateManager(t *testing.T) {
	env, _ := NewMockedEnv(t, false)
	node := env.NewMockedNode("n0")
	time.Sleep(200 * time.Millisecond)
	require.Nil(t, node.StateManager.(*stateManager).solidState)
	require.True(t, len(node.StateManager.(*stateManager).blockCandidates) == 1)
}

func TestGetInitialState(t *testing.T) {
	env, originTx := NewMockedEnv(t, false)
	node := env.NewMockedNode("n0")
	time.Sleep(200 * time.Millisecond)
	require.Nil(t, node.StateManager.(*stateManager).solidState)
	require.True(t, len(node.StateManager.(*stateManager).blockCandidates) == 1)

	originOut, err := utxoutil.GetSingleChainedAliasOutput(originTx)
	require.NoError(t, err)

	env.AddNode(node)
	manager := node.StateManager.(*stateManager)
	manager.EventTimerMsg(2)

	time.Sleep(200 * time.Millisecond)
	require.True(t, originOut.Compare(manager.stateOutput) == 0)
	require.True(t, manager.stateOutput.GetStateIndex() == 0)
	require.EqualValues(t, manager.solidState.Hash(), state.OriginStateHash())
}

func TestGetNextState(t *testing.T) {
	env, originTx := NewMockedEnv(t, false)
	node := env.NewMockedNode("n0")
	time.Sleep(200 * time.Millisecond)
	require.Nil(t, node.StateManager.(*stateManager).solidState)
	require.True(t, len(node.StateManager.(*stateManager).blockCandidates) == 1)

	originOut, err := utxoutil.GetSingleChainedAliasOutput(originTx)
	require.NoError(t, err)

	env.AddNode(node)
	manager := node.StateManager.(*stateManager)
	manager.EventTimerMsg(2)

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
