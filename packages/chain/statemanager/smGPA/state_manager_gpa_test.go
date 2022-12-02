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

// Single node network. 8 blocks are sent to state manager. The result is checked
// by sending consensus requests, which force the access of the blocks.
func TestBasic(t *testing.T) {
	nodeIDs := gpa.MakeTestNodeIDs("Node", 1)
	env := newTestEnv(t, nodeIDs, smGPAUtils.NewMockedBlockWAL)
	defer env.finalize()

	nodeID := nodeIDs[0]
	blocks, stateOutputs := env.bf.GetBlocks(t, 8, 1)
	env.sendBlocksToNode(nodeID, blocks...)

	cspInput, cspRespChan := smInputs.NewConsensusStateProposal(context.Background(), stateOutputs[7])
	env.tc.WithInputs(map[gpa.NodeID]gpa.Input{nodeID: cspInput}).RunAll()
	require.NoError(t, env.requireReceiveAnything(cspRespChan, 5*time.Second))
	commitment, err := state.L1CommitmentFromBytes(stateOutputs[7].GetAliasOutput().StateMetadata)
	require.NoError(t, err)
	cdsInput, cdsRespChan := smInputs.NewConsensusDecidedState(context.Background(), stateOutputs[7])
	env.tc.WithInputs(map[gpa.NodeID]gpa.Input{nodeID: cdsInput}).RunAll()
	require.NoError(t, env.requireReceiveState(cdsRespChan, 8, commitment, 5*time.Second))
}

// 10 nodes in a network. 8 blocks are sent to state manager of the first node.
// The result is checked by sending consensus requests to all the other 9 nodes,
// which force the access (and retrieval) of the blocks. For successful retrieval,
// several timer events are required for nodes to try to request blocks from peers.
func TestManyNodes(t *testing.T) {
	nodeIDs := gpa.MakeTestNodeIDs("Node", 10)
	smTimers := NewStateManagerTimers()
	smTimers.StateManagerGetBlockRetry = 100 * time.Millisecond
	env := newTestEnv(t, nodeIDs, smGPAUtils.NewMockedBlockWAL, smTimers)
	defer env.finalize()

	blocks, stateOutputs := env.bf.GetBlocks(t, 16, 1)
	env.sendBlocksToNode(nodeIDs[0], blocks...)

	// Nodes are checked sequentially
	var result bool
	for i := 1; i < len(nodeIDs); i++ {
		cspInput, cspRespChan := smInputs.NewConsensusStateProposal(context.Background(), stateOutputs[7])
		env.tc.WithInputs(map[gpa.NodeID]gpa.Input{nodeIDs[i]: cspInput}).RunAll()
		env.t.Logf("Sequential: waiting for blocks ending with %s to be available on node %s...", blocks[7].L1Commitment(), nodeIDs[i])
		result = env.requireReceiveAnythingNTimes(cspRespChan, 10, 200*time.Millisecond)
		require.True(t, result)
		commitment, err := state.L1CommitmentFromBytes(stateOutputs[7].GetAliasOutput().StateMetadata)
		require.NoError(t, err)
		cdsInput, cdsRespChan := smInputs.NewConsensusDecidedState(context.Background(), stateOutputs[7])
		env.tc.WithInputs(map[gpa.NodeID]gpa.Input{nodeIDs[i]: cdsInput}).RunAll()
		require.NoError(env.t, env.requireReceiveState(cdsRespChan, 8, commitment, 5*time.Second))
	}
	// Nodes are checked in parallel
	cspInputs := make(map[gpa.NodeID]gpa.Input)
	cspRespChans := make(map[gpa.NodeID]<-chan interface{})
	for i := 1; i < len(nodeIDs); i++ {
		nodeID := nodeIDs[i]
		cspInputs[nodeID], cspRespChans[nodeID] = smInputs.NewConsensusStateProposal(context.Background(), stateOutputs[15])
	}
	env.tc.WithInputs(cspInputs).RunAll()
	for nodeID, cspRespChan := range cspRespChans {
		env.t.Logf("Parallel: waiting for blocks ending with %s to be available on node %s...", blocks[15].L1Commitment(), nodeID)
		result = env.requireReceiveAnythingNTimes(cspRespChan, 10, 200*time.Millisecond)
		require.True(t, result)
	}
	commitment, err := state.L1CommitmentFromBytes(stateOutputs[15].GetAliasOutput().StateMetadata)
	require.NoError(t, err)
	cdsInputs := make(map[gpa.NodeID]gpa.Input)
	cdsRespChans := make(map[gpa.NodeID]<-chan state.State)
	for i := 1; i < len(nodeIDs); i++ {
		nodeID := nodeIDs[i]
		cdsInputs[nodeID], cdsRespChans[nodeID] = smInputs.NewConsensusDecidedState(context.Background(), stateOutputs[15])
	}
	env.tc.WithInputs(cdsInputs).RunAll()
	for nodeID, cdsRespChan := range cdsRespChans {
		env.t.Logf("Parallel: waiting for state %s on node %s", commitment, nodeID)
		require.NoError(env.t, env.requireReceiveState(cdsRespChan, 16, commitment, 5*time.Second))
	}
}

// 12 nodes setting.
// 1. This is repeated 3 times, resulting in 3 consecutive batches of blocks:
//   1.1 Chain of 10 blocks are generated; each of them are sent to a random node
//   1.2 A randomly chosen block in a batch is chosen and is approved
//   1.3 A successful change of state of state manager is waited for; each check
//       fires a timer event to force the exchange of blocks between nodes.
// 2. The last block of the last batch is approved and 1.3 is repeated.
// 3. A random block is chosen in the second batch and 1.1, 1.2, 1.3 and 2 are
//    repeated branching from this block
// 4. A common ancestor (mempool) request is sent to state manager for first and
//    second batch ends.
func TestFull(t *testing.T) {
	nodeCount := 12
	iterationSize := 10
	iterationCount := 3
	maxRetriesPerIteration := 100

	nodeIDs := gpa.MakeTestNodeIDs("Node", nodeCount)
	smTimers := NewStateManagerTimers()
	smTimers.StateManagerGetBlockRetry = 100 * time.Millisecond
	env := newTestEnv(t, nodeIDs, smGPAUtils.NewMockedBlockWAL, smTimers)
	defer env.finalize()

	// artifficialTime := smGPAUtils.NewArtifficialTimeProvider()
	// smTimers.TimeProvider = artifficialTime

	lastAliasOutput := env.bf.GetOriginOutput(t)
	lastCommitment := state.OriginL1Commitment()

	confirmAndWaitFun := func(stateOutput *isc.AliasOutputWithID) bool {
		env.sendInputToNodes(func() gpa.Input {
			return smInputs.NewChainReceiveConfirmedAliasOutput(stateOutput)
		})

		for i := 0; i < maxRetriesPerIteration; i++ {
			t.Logf("\twaiting for approval to propagate through nodes: %v", i)
			if env.isAllNodesAtState(stateOutput) {
				return true
			}
			env.sendTimerTickToNodes(smTimers.StateManagerGetBlockRetry)
		}
		return env.isAllNodesAtState(stateOutput)
	}

	testIterationFun := func(i int, baseAliasOutput *isc.AliasOutputWithID, baseCommitment *state.L1Commitment, incrementFactor ...uint64) ([]*isc.AliasOutputWithID, []*state.L1Commitment) {
		env.t.Logf("Iteration %v: generating %v blocks and sending them to nodes", i, iterationSize)
		blocks, aliasOutputs := env.bf.GetBlocksFrom(t, iterationSize, 1, baseCommitment, baseAliasOutput, incrementFactor...)
		commitments := make([]*state.L1Commitment, len(blocks))
		for j, block := range blocks {
			env.sendBlocksToNode(nodeIDs[rand.Intn(nodeCount)], block)
			commitments[j] = block.L1Commitment()
		}
		confirmOutput := aliasOutputs[rand.Intn(iterationSize-1)] // Do not confirm the last state/blocks
		t.Logf("Iteration %v: approving output index %v, id %v", i, confirmOutput.GetStateIndex(), confirmOutput)
		require.True(t, confirmAndWaitFun(confirmOutput))
		return aliasOutputs, commitments
	}

	var branchAliasOutput *isc.AliasOutputWithID
	var branchCommitment *state.L1Commitment
	oldCommitments := make([]*state.L1Commitment, 0)
	for i := 0; i < iterationCount; i++ {
		aliasOutputs, commitments := testIterationFun(i, lastAliasOutput, lastCommitment, 1)
		lastAliasOutput = aliasOutputs[iterationSize-1]
		lastCommitment = commitments[iterationSize-1]
		if i == 1 {
			branchIndex := rand.Intn(iterationSize)
			branchAliasOutput = aliasOutputs[branchIndex]
			branchCommitment = commitments[branchIndex]
			oldCommitments = append(oldCommitments, commitments[branchIndex+1:]...)
		} else if i == 2 {
			oldCommitments = append(oldCommitments, commitments...)
		}
	}
	require.True(t, confirmAndWaitFun(lastAliasOutput))
	oldAliasOutput := lastAliasOutput

	// Branching from the middle of the second iteration
	require.NotNil(t, branchAliasOutput)
	lastAliasOutput = branchAliasOutput
	lastCommitment = branchCommitment
	newCommitments := make([]*state.L1Commitment, 0)
	for i := 0; i < iterationCount; i++ {
		aliasOutputs, commitments := testIterationFun(i, lastAliasOutput, lastCommitment, 2)
		lastAliasOutput = aliasOutputs[iterationSize-1]
		lastCommitment = commitments[iterationSize-1]
		newCommitments = append(newCommitments, commitments...)
	}
	require.True(t, confirmAndWaitFun(lastAliasOutput))
	newAliasOutput := lastAliasOutput

	// Common ancestor request
	for _, nodeID := range env.nodeIDs {
		request, responseCh := smInputs.NewMempoolStateRequest(context.Background(), oldAliasOutput, newAliasOutput)
		env.t.Logf("Requesting state for Mempool from node %s", nodeID)
		env.tc.WithInputs(map[gpa.NodeID]gpa.Input{nodeID: request}).RunAll()
		require.NoError(env.t, env.requireReceiveMempoolResults(responseCh, oldCommitments, newCommitments, 5*time.Second))
	}
}

// Single node network. Checks if block cache is cleaned via state manager
// timer events.
func TestBlockCacheCleaningAuto(t *testing.T) {
	nodeIDs := gpa.MakeTestNodeIDs("Node", 1)
	smTimers := NewStateManagerTimers()
	smTimers.BlockCacheBlocksInCacheDuration = 300 * time.Millisecond
	smTimers.BlockCacheBlockCleaningPeriod = 70 * time.Millisecond
	env := newTestEnv(t, nodeIDs, smGPAUtils.NewMockedBlockWAL, smTimers)
	defer env.finalize()

	nodeID := nodeIDs[0]
	blocks, _ := env.bf.GetBlocks(t, 6, 2)

	blockCache := env.sms[nodeID].(*stateManagerGPA).blockCache
	blockCache.AddBlock(blocks[0])
	blockCache.AddBlock(blocks[1])
	require.NotNil(env.t, blockCache.GetBlock(blocks[0].L1Commitment()))
	require.NotNil(env.t, blockCache.GetBlock(blocks[1].L1Commitment()))
	env.sendTimerTickToNodes(100 * time.Millisecond)
	require.NotNil(env.t, blockCache.GetBlock(blocks[0].L1Commitment()))
	require.NotNil(env.t, blockCache.GetBlock(blocks[1].L1Commitment()))
	blockCache.AddBlock(blocks[2])
	require.NotNil(env.t, blockCache.GetBlock(blocks[0].L1Commitment()))
	require.NotNil(env.t, blockCache.GetBlock(blocks[1].L1Commitment()))
	require.NotNil(env.t, blockCache.GetBlock(blocks[2].L1Commitment()))
	env.sendTimerTickToNodes(100 * time.Millisecond)
	require.NotNil(env.t, blockCache.GetBlock(blocks[0].L1Commitment()))
	require.NotNil(env.t, blockCache.GetBlock(blocks[1].L1Commitment()))
	require.NotNil(env.t, blockCache.GetBlock(blocks[2].L1Commitment()))
	blockCache.AddBlock(blocks[3])
	require.NotNil(env.t, blockCache.GetBlock(blocks[0].L1Commitment()))
	require.NotNil(env.t, blockCache.GetBlock(blocks[1].L1Commitment()))
	require.NotNil(env.t, blockCache.GetBlock(blocks[2].L1Commitment()))
	require.NotNil(env.t, blockCache.GetBlock(blocks[3].L1Commitment()))
	env.sendTimerTickToNodes(80 * time.Millisecond)
	require.NotNil(env.t, blockCache.GetBlock(blocks[0].L1Commitment()))
	require.NotNil(env.t, blockCache.GetBlock(blocks[1].L1Commitment()))
	require.NotNil(env.t, blockCache.GetBlock(blocks[2].L1Commitment()))
	require.NotNil(env.t, blockCache.GetBlock(blocks[3].L1Commitment()))
	env.sendTimerTickToNodes(100 * time.Millisecond)
	blockCache.AddBlock(blocks[4])
	require.Nil(env.t, blockCache.GetBlock(blocks[0].L1Commitment()))
	require.Nil(env.t, blockCache.GetBlock(blocks[1].L1Commitment()))
	require.NotNil(env.t, blockCache.GetBlock(blocks[2].L1Commitment()))
	require.NotNil(env.t, blockCache.GetBlock(blocks[3].L1Commitment()))
	require.NotNil(env.t, blockCache.GetBlock(blocks[4].L1Commitment()))
	env.sendTimerTickToNodes(100 * time.Millisecond)
	require.Nil(env.t, blockCache.GetBlock(blocks[2].L1Commitment()))
	require.NotNil(env.t, blockCache.GetBlock(blocks[3].L1Commitment()))
	require.NotNil(env.t, blockCache.GetBlock(blocks[4].L1Commitment()))
	env.sendTimerTickToNodes(100 * time.Millisecond)
	require.Nil(env.t, blockCache.GetBlock(blocks[3].L1Commitment()))
	require.NotNil(env.t, blockCache.GetBlock(blocks[4].L1Commitment()))
	env.sendTimerTickToNodes(200 * time.Millisecond)
	require.Nil(env.t, blockCache.GetBlock(blocks[4].L1Commitment()))
}

func newTestEnv(t *testing.T, nodeIDs []gpa.NodeID, createWALFun func() smGPAUtils.BlockWAL, timersOpt ...StateManagerTimers) *testEnv {
	bf := smGPAUtils.NewBlockFactory()
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
		cbpInput, cbpRespChan := smInputs.NewConsensusBlockProduced(context.Background(), teT.bf.GetStateDraft(teT.t, blocks[i]))
		teT.t.Logf("Supplying block %s to node %s", blocks[i].L1Commitment(), nodeID)
		teT.tc.WithInputs(map[gpa.NodeID]gpa.Input{nodeID: cbpInput}).RunAll()
		require.NoError(teT.t, teT.requireReceiveNoError(cbpRespChan, 5*time.Second))
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
		require.True(teT.t, commitment.GetTrieRoot().Equals(s.TrieRoot()))
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("Waiting to receive state timeouted")
	}
}

func (teT *testEnv) requireReceiveMempoolResults(respChan <-chan *smInputs.MempoolStateRequestResults, oldCommitments, newCommitments []*state.L1Commitment, timeout time.Duration) error {
	select {
	case msrr := <-respChan:
		require.True(teT.t, msrr.GetNewState().TrieRoot().Equals(newCommitments[len(newCommitments)-1].GetTrieRoot()))
		requireEqualsFun := func(expectedCommitments []*state.L1Commitment, received []state.Block) {
			require.Equal(teT.t, len(expectedCommitments), len(received))
			for i := range expectedCommitments {
				receivedCommitment := received[i].L1Commitment()
				teT.t.Logf("\tchecking %v-th element: expected %s, received %s", i, expectedCommitments[i], receivedCommitment)
				require.True(teT.t, expectedCommitments[i].Equals(receivedCommitment))
			}
		}
		teT.t.Logf("Checking added blocks...")
		requireEqualsFun(newCommitments, msrr.GetAdded())
		teT.t.Logf("Checking removed blocks...")
		requireEqualsFun(oldCommitments, msrr.GetRemoved())
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("Waiting to receive mempool results timeouted")
	}
}

func (teT *testEnv) sendTimerTickToNodes(delay time.Duration) {
	now := teT.timeProvider.GetNow().Add(delay)
	teT.timeProvider.SetNow(now)
	teT.t.Logf("Time %v is sent to nodes %v", now, teT.nodeIDs)
	teT.sendInputToNodes(func() gpa.Input {
		return smInputs.NewStateManagerTimerTick(now)
	})
}

func (teT *testEnv) sendInputToNodes(makeInputFun func() gpa.Input) {
	inputs := make(map[gpa.NodeID]gpa.Input)
	for _, nodeID := range teT.nodeIDs {
		inputs[nodeID] = makeInputFun()
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
		if !expectedCommitment.GetTrieRoot().Equals(sm.currentL1Commitment.GetTrieRoot()) {
			teT.t.Logf("Node %s is at state index %v, but state commitments do not match: expected %s, obtained %s",
				nodeID, stateOutput.GetStateIndex(), expectedCommitment.GetTrieRoot(), sm.currentL1Commitment.GetTrieRoot())
			return false
		}
		if !expectedCommitment.GetBlockHash().Equals(sm.currentL1Commitment.GetBlockHash()) {
			teT.t.Logf("Node %s is at state index %v, but block hashes do not match: expected %s, obtained %s",
				nodeID, stateOutput.GetStateIndex(), expectedCommitment.GetBlockHash(), sm.currentL1Commitment.GetBlockHash())
			return false
		}
	}
	return true
}
