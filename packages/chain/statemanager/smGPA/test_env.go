package smGPA

import (
	"context"
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
	_ = teT.log.Sync()
}

func (teT *testEnv) sendBlocksToNode(nodeID gpa.NodeID, timeStep time.Duration, blocks ...state.Block) {
	// If `ConsensusBlockProduced` is sent to the node, the node has definitely obtained all the blocks
	// needed to commit this block. This is ensured by consensus.
	require.True(teT.t, teT.sendAndEnsureCompletedConsensusStateProposal(blocks[0].PreviousL1Commitment(), nodeID, 100, timeStep))
	for i := range blocks {
		teT.t.Logf("Supplying block %s to node %s", blocks[i].L1Commitment(), nodeID)
		teT.sendAndEnsureCompletedConsensusBlockProduced(blocks[i], nodeID, 100, timeStep)
	}
}

func (teT *testEnv) sendBlocksToRandomNode(nodeIDs []gpa.NodeID, timeStep time.Duration, blocks ...state.Block) {
	for _, block := range blocks {
		teT.sendBlocksToNode(nodeIDs[rand.Intn(len(nodeIDs))], timeStep, block)
	}
}

// --------

func (teT *testEnv) sendAndEnsureCompletedConsensusBlockProduced(block state.Block, nodeID gpa.NodeID, maxTimeIterations int, timeStep time.Duration) bool {
	responseCh := teT.sendConsensusBlockProduced(block, nodeID)
	return teT.ensureCompletedConsensusBlockProduced(responseCh, maxTimeIterations, timeStep)
}

func (teT *testEnv) sendConsensusBlockProduced(block state.Block, nodeID gpa.NodeID) <-chan error {
	input, responseCh := smInputs.NewConsensusBlockProduced(context.Background(), teT.bf.GetStateDraft(block))
	teT.tc.WithInputs(map[gpa.NodeID]gpa.Input{nodeID: input}).RunAll()
	return responseCh
}

func (teT *testEnv) ensureCompletedConsensusBlockProduced(respChan <-chan error, maxTimeIterations int, timeStep time.Duration) bool {
	return teT.ensureTrue("response from ConsensusBlockProduced", func() bool {
		select {
		case err := <-respChan:
			require.NoError(teT.t, err)
			return true
		default:
			return false
		}
	}, maxTimeIterations, timeStep)
}

// --------

func (teT *testEnv) sendAndEnsureCompletedConsensusStateProposal(commitment *state.L1Commitment, nodeID gpa.NodeID, maxTimeIterations int, timeStep time.Duration) bool {
	responseCh := teT.sendConsensusStateProposal(commitment, nodeID)
	return teT.ensureCompletedConsensusStateProposal(responseCh, maxTimeIterations, timeStep)
}

func (teT *testEnv) sendConsensusStateProposal(commitment *state.L1Commitment, nodeID gpa.NodeID) <-chan interface{} {
	input, responseCh := smInputs.NewConsensusStateProposal(context.Background(), teT.bf.GetAliasOutput(commitment))
	teT.tc.WithInputs(map[gpa.NodeID]gpa.Input{nodeID: input}).RunAll()
	return responseCh
}

func (teT *testEnv) ensureCompletedConsensusStateProposal(respChan <-chan interface{}, maxTimeIterations int, timeStep time.Duration) bool {
	return teT.ensureTrue("response from ConsensusStateProposal", func() bool {
		select {
		case result := <-respChan:
			require.Nil(teT.t, result)
			return true
		default:
			return false
		}
	}, maxTimeIterations, timeStep)
}

// --------

func (teT *testEnv) sendAndEnsureCompletedConsensusDecidedState(commitment *state.L1Commitment, nodeID gpa.NodeID, maxTimeIterations int, timeStep time.Duration) bool {
	responseCh := teT.sendConsensusDecidedState(commitment, nodeID)
	return teT.ensureCompletedConsensusDecidedState(responseCh, commitment, maxTimeIterations, timeStep)
}

func (teT *testEnv) sendConsensusDecidedState(commitment *state.L1Commitment, nodeID gpa.NodeID) <-chan state.State {
	input, responseCh := smInputs.NewConsensusDecidedState(context.Background(), teT.bf.GetAliasOutput(commitment))
	teT.tc.WithInputs(map[gpa.NodeID]gpa.Input{nodeID: input}).RunAll()
	return responseCh
}

func (teT *testEnv) ensureCompletedConsensusDecidedState(respChan <-chan state.State, expectedCommitment *state.L1Commitment, maxTimeIterations int, timeStep time.Duration) bool {
	expectedState := teT.bf.GetState(expectedCommitment)
	return teT.ensureTrue("response from ConsensusDecidedState", func() bool {
		select {
		case s := <-respChan:
			// Should be require.True(teT.t, expected.Equals(s))
			expectedTrieRoot := expectedState.TrieRoot()
			receivedTrieRoot := s.TrieRoot()
			require.Equal(teT.t, expectedState.BlockIndex(), s.BlockIndex())
			teT.t.Logf("Checking trie roots: expected %s, obtained %s", expectedTrieRoot, receivedTrieRoot)
			require.True(teT.t, expectedTrieRoot.Equals(receivedTrieRoot))
			return true
		default:
			return false
		}
	}, maxTimeIterations, timeStep)
}

// --------

func (teT *testEnv) sendAndEnsureCompletedChainFetchStateDiff(oldCommitment, newCommitment *state.L1Commitment, expectedOldBlocks, expectedNewBlocks []state.Block, nodeID gpa.NodeID, maxTimeIterations int, timeStep time.Duration) bool {
	responseCh := teT.sendChainFetchStateDiff(oldCommitment, newCommitment, nodeID)
	return teT.ensureCompletedChainFetchStateDiff(responseCh, expectedOldBlocks, expectedNewBlocks, maxTimeIterations, timeStep)
}

func (teT *testEnv) sendChainFetchStateDiff(oldCommitment, newCommitment *state.L1Commitment, nodeID gpa.NodeID) <-chan *smInputs.ChainFetchStateDiffResults {
	input, responseCh := smInputs.NewChainFetchStateDiff(context.Background(), teT.bf.GetAliasOutput(oldCommitment), teT.bf.GetAliasOutput(newCommitment))
	teT.tc.WithInputs(map[gpa.NodeID]gpa.Input{nodeID: input}).RunAll()
	return responseCh
}

func (teT *testEnv) ensureCompletedChainFetchStateDiff(respChan <-chan *smInputs.ChainFetchStateDiffResults, expectedOldBlocks, expectedNewBlocks []state.Block, maxTimeIterations int, timeStep time.Duration) bool {
	return teT.ensureTrue("response from ChainFetchStateDiff", func() bool {
		select {
		case cfsdr := <-respChan:
			newStateTrieRoot := cfsdr.GetNewState().TrieRoot()
			lastNewBlockTrieRoot := expectedNewBlocks[len(expectedNewBlocks)-1].TrieRoot()
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
			teT.t.Log("Checking added blocks...")
			requireEqualsFun(expectedNewBlocks, cfsdr.GetAdded())
			teT.t.Log("Checking removed blocks...")
			requireEqualsFun(expectedOldBlocks, cfsdr.GetRemoved())
			return true
		default:
			return false
		}
	}, maxTimeIterations, timeStep)
}

// --------

func (teT *testEnv) ensureTrue(title string, predicate func() bool, maxTimeIterations int, timeStep time.Duration) bool {
	for i := 0; i < maxTimeIterations; i++ {
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
