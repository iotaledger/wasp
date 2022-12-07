package smGPA

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/kvstore/mapdb"
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smGPAUtils"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smInputs"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smUtils"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

type testEnv struct {
	t            *testing.T
	bf           *smGPAUtils.BlockFactory
	nodeIDs      []gpa.NodeID
	timeProvider smGPAUtils.TimeProvider
	sms          map[gpa.NodeID]gpa.GPA
	tc           *gpa.TestContext
	log          *logger.Logger
}

func newTestEnv(t *testing.T, nodeIDs []gpa.NodeID, createWALFun func() smGPAUtils.BlockWAL, timersOpt ...StateManagerTimers) *testEnv {
	bf := smGPAUtils.NewBlockFactory(t)
	chainID := bf.GetChainID()
	log := testlogger.NewLogger(t).Named("c-" + chainID.ShortString())
	sms := make(map[gpa.NodeID]gpa.GPA)
	var timers StateManagerTimers
	if len(timersOpt) > 0 {
		timers = timersOpt[0]
	} else {
		timers = NewStateManagerTimers()
	}
	timers.TimeProvider = smGPAUtils.NewArtifficialTimeProvider()
	for _, nodeID := range nodeIDs {
		var err error
		smLog := log.Named(nodeID.String())
		nr := smUtils.NewNodeRandomiser(nodeID, nodeIDs, smLog)
		wal := createWALFun()
		store := state.InitChainStore(mapdb.NewMapDB())
		sms[nodeID], err = New(chainID, nr, wal, store, smLog, timers)
		require.NoError(t, err)
	}
	return &testEnv{
		t:            t,
		bf:           bf,
		nodeIDs:      nodeIDs,
		timeProvider: timers.TimeProvider,
		sms:          sms,
		tc:           gpa.NewTestContext(sms),
		log:          log,
	}
}

func (teT *testEnv) finalize() {
	teT.log.Sync()
}

func (teT *testEnv) sendBlocksToNode(nodeID gpa.NodeID, blocks ...state.Block) {
	for i := range blocks {
		cbpInput, cbpRespChan := smInputs.NewConsensusBlockProduced(context.Background(), teT.bf.GetStateDraft(blocks[i]))
		teT.t.Logf("Supplying block %s to node %s", blocks[i].L1Commitment(), nodeID)
		teT.tc.WithInputs(map[gpa.NodeID]gpa.Input{nodeID: cbpInput}).RunAll()
		require.NoError(teT.t, teT.requireReceiveNoError(cbpRespChan, 5*time.Second))
	}
}

func (teT *testEnv) sendBlocksToRandomNode(nodeIDs []gpa.NodeID, blocks ...state.Block) {
	for _, block := range blocks {
		teT.sendBlocksToNode(nodeIDs[rand.Intn(len(nodeIDs))], block)
	}
}

func (teT *testEnv) requireReceiveAnything(anyChan <-chan (interface{}), timeout time.Duration) error { //nolint:gocritic
	select {
	case <-anyChan:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("Waiting to receive anything timeouted")
	}
}

func (teT *testEnv) requireReceiveAnythingNTimes(anyChan <-chan interface{}, n int, delay time.Duration) bool {
	for j := 0; j < n; j++ {
		teT.t.Logf("\t...iteration %v", j)
		if teT.requireReceiveAnything(anyChan, 0*time.Second) == nil {
			return true
		}
		teT.sendTimerTickToNodes(delay)
	}
	return false
}

func (teT *testEnv) requireReceiveNoError(errChan <-chan (error), timeout time.Duration) error { //nolint:gocritic
	select {
	case err := <-errChan:
		require.NoError(teT.t, err)
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("Waiting to receive no error timeouted")
	}
}

func (teT *testEnv) requireReceiveState(respChan <-chan state.State, index uint32, commitment *state.L1Commitment, timeout time.Duration) error {
	select {
	case s := <-respChan:
		require.Equal(teT.t, s.BlockIndex(), index)
		require.True(teT.t, commitment.TrieRoot().Equals(s.TrieRoot()))
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("Waiting to receive state timeouted")
	}
}

func (teT *testEnv) requireReceiveMempoolResults(respChan <-chan *smInputs.MempoolStateRequestResults, oldBlocks, newBlocks []state.Block, timeout time.Duration) error {
	select {
	case msrr := <-respChan:
		newStateTrieRoot := msrr.GetNewState().TrieRoot()
		lastNewBlockTrieRoot := newBlocks[len(newBlocks)-1].TrieRoot()
		teT.t.Logf("Checking trie roots: expected %s, obtained %s", lastNewBlockTrieRoot, newStateTrieRoot)
		require.True(teT.t, newStateTrieRoot.Equals(lastNewBlockTrieRoot))
		requireEqualsFun := func(expected, received []state.Block) {
			teT.t.Logf("\tExpected %v elements, obtained %v elements", len(expected), len(received))
			require.Equal(teT.t, len(expected), len(received))
			for i := range expected {
				expectedCommitment := expected[i].L1Commitment()
				receivedCommitment := received[i].L1Commitment()
				teT.t.Logf("\tchecking %v-th element: expected %s, received %s", i, expectedCommitment, receivedCommitment)
				require.True(teT.t, expectedCommitment.Equals(receivedCommitment))
			}
		}
		teT.t.Logf("Checking added blocks...")
		requireEqualsFun(newBlocks, msrr.GetAdded())
		teT.t.Logf("Checking removed blocks...")
		requireEqualsFun(oldBlocks, msrr.GetRemoved())
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("Waiting to receive mempool results timeouted")
	}
}

func (teT *testEnv) requireAfterTime(title string, predicate func() bool, maxTime int, timeStep time.Duration) bool {
	for i := 0; i < maxTime; i++ {
		teT.t.Logf("Waiting for %s iteration %v", title, i)
		if predicate() {
			return true
		}
		teT.sendTimerTickToNodes(timeStep)
	}
	return false
}

func (teT *testEnv) sendTimerTickToNodes(delay time.Duration) {
	now := teT.timeProvider.GetNow().Add(delay)
	teT.timeProvider.SetNow(now)
	teT.t.Logf("Time %v is sent to nodes %v", now, teT.nodeIDs)
	teT.sendInputToNodes(func(_ gpa.NodeID) gpa.Input {
		return smInputs.NewStateManagerTimerTick(now)
	})
}

func (teT *testEnv) sendInputToNodes(makeInputFun func(gpa.NodeID) gpa.Input) {
	inputs := make(map[gpa.NodeID]gpa.Input)
	for _, nodeID := range teT.nodeIDs {
		inputs[nodeID] = makeInputFun(nodeID)
	}
	teT.tc.WithInputs(inputs).RunAll()
}

func (teT *testEnv) isAllNodesAtState(stateOutput *isc.AliasOutputWithID) bool {
	for nodeID, smGPA := range teT.sms {
		sm, ok := smGPA.(*stateManagerGPA)
		require.True(teT.t, ok)
		expectedCommitment, err := state.L1CommitmentFromAliasOutput(stateOutput.GetAliasOutput())
		require.NoError(teT.t, err)
		if stateOutput.GetStateIndex() != sm.currentStateIndex {
			teT.t.Logf("Node %s is not yet at state index %v, it is at state index %v",
				nodeID, stateOutput.GetStateIndex(), sm.currentStateIndex)
			return false
		}
		if !expectedCommitment.TrieRoot().Equals(sm.currentL1Commitment.TrieRoot()) {
			teT.t.Logf("Node %s is at state index %v, but state commitments do not match: expected %s, obtained %s",
				nodeID, stateOutput.GetStateIndex(), expectedCommitment.TrieRoot(), sm.currentL1Commitment.TrieRoot())
			return false
		}
		if !expectedCommitment.BlockHash().Equals(sm.currentL1Commitment.BlockHash()) {
			teT.t.Logf("Node %s is at state index %v, but block hashes do not match: expected %s, obtained %s",
				nodeID, stateOutput.GetStateIndex(), expectedCommitment.BlockHash(), sm.currentL1Commitment.BlockHash())
			return false
		}
	}
	return true
}
