package smGPA

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smGPAUtils"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smInputs"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/state"
)

// Single node network. 8 blocks are sent to state manager. The result is checked
// by checking the store and sending consensus requests, which force the access
// of the blocks.
func TestBasic(t *testing.T) {
	nodeIDs := gpa.MakeTestNodeIDs(1)
	env := newTestEnv(t, nodeIDs, smGPAUtils.NewMockedBlockWAL)
	defer env.finalize()

	nodeID := nodeIDs[0]
	blocks := env.bf.GetBlocks(8, 1)
	env.sendBlocksToNode(nodeID, 0*time.Second, blocks...)
	require.True(env.t, env.ensureStoreContainsBlocksNoWait(nodeID, blocks))

	lastCommitment := blocks[7].L1Commitment()
	require.True(env.t, env.sendAndEnsureCompletedConsensusStateProposal(lastCommitment, nodeID, 1, 0*time.Second))
	require.True(env.t, env.sendAndEnsureCompletedConsensusDecidedState(lastCommitment, nodeID, 1, 0*time.Second))
}

// 10 nodes in a network. 8 blocks are sent to state manager of the first node.
// The result is checked by sending consensus requests to all the other 9 nodes,
// which force the access (and retrieval) of the blocks. For successful retrieval,
// several timer events are required for nodes to try to request blocks from peers.
func TestManyNodes(t *testing.T) {
	nodeIDs := gpa.MakeTestNodeIDs(10)
	smTimers := NewStateManagerTimers()
	smTimers.StateManagerGetBlockRetry = 100 * time.Millisecond
	env := newTestEnv(t, nodeIDs, smGPAUtils.NewMockedBlockWAL, smTimers)
	defer env.finalize()

	blocks := env.bf.GetBlocks(16, 1)
	env.sendBlocksToNode(nodeIDs[0], smTimers.StateManagerGetBlockRetry, blocks...)
	require.True(env.t, env.ensureStoreContainsBlocksNoWait(nodeIDs[0], blocks))

	// Nodes are checked sequentially
	lastCommitment := blocks[7].L1Commitment()
	for i := 1; i < len(nodeIDs); i++ {
		env.t.Logf("Sequential: waiting for blocks ending with %s to be available on node %s...", lastCommitment, nodeIDs[i].ShortString())
		require.True(env.t, env.sendAndEnsureCompletedConsensusStateProposal(lastCommitment, nodeIDs[i], 100, smTimers.StateManagerGetBlockRetry))
		require.True(env.t, env.sendAndEnsureCompletedConsensusDecidedState(lastCommitment, nodeIDs[i], 8, smTimers.StateManagerGetBlockRetry))
		require.True(env.t, env.ensureStoreContainsBlocksNoWait(nodeIDs[i], blocks[:8]))
	}
	// Nodes are checked in parallel
	lastCommitment = blocks[15].L1Commitment()
	lastOutput := env.bf.GetAliasOutput(lastCommitment)
	cspInputs := make(map[gpa.NodeID]gpa.Input)
	cspRespChans := make(map[gpa.NodeID]<-chan interface{})
	for i := 1; i < len(nodeIDs); i++ {
		nodeID := nodeIDs[i]
		cspInputs[nodeID], cspRespChans[nodeID] = smInputs.NewConsensusStateProposal(context.Background(), lastOutput)
	}
	env.tc.WithInputs(cspInputs).RunAll()
	for nodeID, cspRespChan := range cspRespChans {
		env.t.Logf("Parallel: waiting for blocks ending with %s to be available on node %s...", lastCommitment, nodeID.ShortString())
		require.True(env.t, env.ensureCompletedConsensusStateProposal(cspRespChan, 10, smTimers.StateManagerGetBlockRetry))
	}
	cdsInputs := make(map[gpa.NodeID]gpa.Input)
	cdsRespChans := make(map[gpa.NodeID]<-chan state.State)
	for i := 1; i < len(nodeIDs); i++ {
		nodeID := nodeIDs[i]
		cdsInputs[nodeID], cdsRespChans[nodeID] = smInputs.NewConsensusDecidedState(context.Background(), lastOutput)
	}
	env.tc.WithInputs(cdsInputs).RunAll()
	for nodeID, cdsRespChan := range cdsRespChans {
		env.t.Logf("Parallel: waiting for state %s on node %s", lastCommitment, nodeID.ShortString())
		require.True(env.t, env.ensureCompletedConsensusDecidedState(cdsRespChan, lastCommitment, 16, smTimers.StateManagerGetBlockRetry))
	}
	for _, nodeID := range nodeIDs {
		env.t.Logf("Parallel: waiting for blocks to be available in store on node %s", nodeID.ShortString())
		require.True(env.t, env.ensureStoreContainsBlocksNoWait(nodeID, blocks[8:]))
	}
}

// 12 nodes setting.
//  1. This is repeated 3 times, resulting in 3 consecutive batches of blocks:
//     1.1 Chain of 10 blocks are generated; each of them are sent to a random node
//     1.2 For each node a randomly chosen block in a batch is chosen and ConsensusDecidedState
//     request is sent.
//     1.3 A response of the request is waited for; each check for the result
//     fires a timer event to force the exchange of blocks between nodes.
//  2. ConsensusDecidedState is sent to each node for the last block of the last
//     batch and 1.3 is repeated.
//  3. A random block is chosen in the second batch and 1.1, 1.2, 1.3 and 2 are
//     repeated branching from this block
//  4. A ChainFetchStateDiff request is sent to state manager for first and
//     second batch ends.
func TestFull(t *testing.T) {
	nodeCount := 12
	iterationSize := 10
	iterationCount := 3
	maxRetriesPerIteration := 100

	nodeIDs := gpa.MakeTestNodeIDs(nodeCount)
	smTimers := NewStateManagerTimers()
	smTimers.StateManagerGetBlockRetry = 100 * time.Millisecond
	env := newTestEnv(t, nodeIDs, smGPAUtils.NewMockedBlockWAL, smTimers)
	defer env.finalize()

	lastCommitment := state.OriginL1Commitment()

	testIterationFun := func(i int, baseCommitment *state.L1Commitment, incrementFactor ...uint64) []state.Block {
		env.t.Logf("Iteration %v: generating %v blocks and sending them to nodes", i, iterationSize)
		blocks := env.bf.GetBlocksFrom(iterationSize, 1, baseCommitment, incrementFactor...)
		env.sendBlocksToRandomNode(nodeIDs, smTimers.StateManagerGetBlockRetry, blocks...)
		for _, nodeID := range nodeIDs {
			randCommitment := blocks[rand.Intn(iterationSize-1)].L1Commitment() // Do not pick the last state/blocks
			t.Logf("Iteration %v: sending ConsensusDecidedState for commitment %s to node %s",
				i, randCommitment, nodeID.ShortString())
			require.True(env.t, env.sendAndEnsureCompletedConsensusDecidedState(randCommitment, nodeID, maxRetriesPerIteration, smTimers.StateManagerGetBlockRetry))
		}
		return blocks
	}

	var branchCommitment *state.L1Commitment
	oldBlocks := make([]state.Block, 0)
	allBlocks := make([]state.Block, 0)
	for i := 0; i < iterationCount; i++ {
		blocks := testIterationFun(i, lastCommitment, 1)
		lastCommitment = blocks[iterationSize-1].L1Commitment()
		if i == 1 {
			branchIndex := rand.Intn(iterationSize)
			branchCommitment = blocks[branchIndex].L1Commitment()
			oldBlocks = append(oldBlocks, blocks[branchIndex+1:]...)
		} else if i == 2 {
			oldBlocks = append(oldBlocks, blocks...)
		}
		allBlocks = append(allBlocks, blocks...)
	}
	for _, nodeID := range nodeIDs {
		t.Logf("Sending ConsensusDecidedState for last original commitment %s to node %s", lastCommitment, nodeID.ShortString())
		require.True(env.t, env.sendAndEnsureCompletedConsensusDecidedState(lastCommitment, nodeID, maxRetriesPerIteration, smTimers.StateManagerGetBlockRetry))
		require.True(env.t, env.ensureStoreContainsBlocksNoWait(nodeID, allBlocks))
	}
	oldCommitment := lastCommitment

	// Branching from the middle of the second iteration
	lastCommitment = branchCommitment
	newBlocks := make([]state.Block, 0)
	t.Logf("Branching from commitment %s", lastCommitment)
	for i := 0; i < iterationCount; i++ {
		blocks := testIterationFun(i, lastCommitment, 2)
		lastCommitment = blocks[iterationSize-1].L1Commitment()
		newBlocks = append(newBlocks, blocks...)
	}
	for _, nodeID := range nodeIDs {
		t.Logf("Sending ConsensusDecidedState for last branch commitment %s to node %s", lastCommitment, nodeID.ShortString())
		require.True(env.t, env.sendAndEnsureCompletedConsensusDecidedState(lastCommitment, nodeID, maxRetriesPerIteration, smTimers.StateManagerGetBlockRetry))
		require.True(env.t, env.ensureStoreContainsBlocksNoWait(nodeID, newBlocks))
	}
	newCommitment := lastCommitment

	// ChainFetchStateDiff request
	for _, nodeID := range env.nodeIDs {
		env.t.Logf("Requesting state for Mempool from node %s", nodeID.ShortString())
		require.True(env.t, env.sendAndEnsureCompletedChainFetchStateDiff(oldCommitment, newCommitment, oldBlocks, newBlocks, nodeID, maxRetriesPerIteration, smTimers.StateManagerGetBlockRetry))
	}
}

// 15 nodes setting.
//  1. A batch of 20 consecutive blocks is generated, each of them is sent to the
//     first node.
//  2. A block is selected at random in the middle (from index 5 to 12), and a
//     branch of 5 blocks is generated. Each of the blocks is sent to the first node.
//  3. For each node a ChainFetchStateDiff request is sent and the successful
//     completion is waited for; each check fires a timer event to force
//     the exchange of blocks between nodes.
func TestMempoolRequest(t *testing.T) {
	nodeCount := 15
	mainSize := 20
	randomFrom := 5
	randomTo := 12
	branchSize := 5
	maxRetries := 100

	nodeIDs := gpa.MakeTestNodeIDs(nodeCount)
	smTimers := NewStateManagerTimers()
	smTimers.StateManagerGetBlockRetry = 100 * time.Millisecond
	env := newTestEnv(t, nodeIDs, smGPAUtils.NewMockedBlockWAL, smTimers)
	defer env.finalize()

	mainBlocks := env.bf.GetBlocks(mainSize, 1)
	branchIndex := randomFrom + rand.Intn(randomTo-randomFrom)
	branchBlocks := env.bf.GetBlocksFrom(branchSize, 1, mainBlocks[branchIndex].L1Commitment(), 2)

	env.sendBlocksToNode(nodeIDs[0], smTimers.StateManagerGetBlockRetry, mainBlocks...)
	require.True(env.t, env.ensureStoreContainsBlocksNoWait(nodeIDs[0], mainBlocks))
	env.sendBlocksToNode(nodeIDs[0], smTimers.StateManagerGetBlockRetry, branchBlocks...)
	require.True(env.t, env.ensureStoreContainsBlocksNoWait(nodeIDs[0], branchBlocks))

	oldCommitment := mainBlocks[len(mainBlocks)-1].L1Commitment()
	newCommitment := branchBlocks[len(branchBlocks)-1].L1Commitment()
	oldBlocks := mainBlocks[branchIndex+1:]
	for _, nodeID := range nodeIDs[1:] {
		env.t.Logf("Sending ChainFetchStateDiff for old commitment %s and new commitment %s to node %s", oldCommitment, newCommitment, nodeID.ShortString())
		require.True(env.t, env.sendAndEnsureCompletedChainFetchStateDiff(oldCommitment, newCommitment, oldBlocks, branchBlocks, nodeID, maxRetries, smTimers.StateManagerGetBlockRetry))
		require.True(env.t, env.ensureStoreContainsBlocksNoWait(nodeID, mainBlocks))
		require.True(env.t, env.ensureStoreContainsBlocksNoWait(nodeID, branchBlocks))
	}
}

// Single node setting.
//  1. A single block is generated and sent to the node.
//  2. A ChainFetchStateDiff request is sent for block 1 as a new block
//     and block 0 as an old block.
func TestMempoolRequestFirstStep(t *testing.T) {
	nodeIDs := gpa.MakeTestNodeIDs(1)
	env := newTestEnv(t, nodeIDs, smGPAUtils.NewMockedBlockWAL)
	defer env.finalize()

	nodeID := nodeIDs[0]
	blocks := env.bf.GetBlocks(1, 1)
	env.sendBlocksToNode(nodeID, 0*time.Second, blocks[0])
	require.True(env.t, env.ensureStoreContainsBlocksNoWait(nodeID, blocks))

	oldCommitment := state.OriginL1Commitment()
	newCommitment := blocks[0].L1Commitment()
	oldBlocks := make([]state.Block, 0)
	require.True(env.t, env.sendAndEnsureCompletedChainFetchStateDiff(oldCommitment, newCommitment, oldBlocks, blocks, nodeID, 1, 0*time.Second))
}

// Single node setting.
//  1. A batch of 10 consecutive blocks is generated, each of them is sent to the node.
//  2. A ChainFetchStateDiff request is sent for block 10 as a new block
//     and block 5 as an old block.
func TestMempoolRequestNoBranch(t *testing.T) {
	batchSize := 10
	middleBlock := 4

	nodeIDs := gpa.MakeTestNodeIDs(1)
	env := newTestEnv(t, nodeIDs, smGPAUtils.NewMockedBlockWAL)
	defer env.finalize()

	nodeID := nodeIDs[0]
	blocks := env.bf.GetBlocks(batchSize, 1)
	env.sendBlocksToNode(nodeID, 0*time.Second, blocks...)
	require.True(env.t, env.ensureStoreContainsBlocksNoWait(nodeID, blocks))

	oldCommitment := blocks[middleBlock].L1Commitment()
	newCommitment := blocks[len(blocks)-1].L1Commitment()
	oldBlocks := make([]state.Block, 0)
	newBlocks := blocks[middleBlock+1:]
	require.True(env.t, env.sendAndEnsureCompletedChainFetchStateDiff(oldCommitment, newCommitment, oldBlocks, newBlocks, nodeID, 1, 0*time.Second))
}

// Single node setting.
//  1. A batch of 10 consecutive blocks is generated, each of them is sent to the node.
//  2. A batch of 8 consecutive blocks is branched from origin commitment. Each of
//     the blocks is sent to the node.
//  3. A ChainFetchStateDiff request is sent for the branch as a new and
//     and original batch as old.
func TestMempoolRequestBranchFromOrigin(t *testing.T) {
	batchSize := 10
	branchSize := 8

	nodeIDs := gpa.MakeTestNodeIDs(1)
	env := newTestEnv(t, nodeIDs, smGPAUtils.NewMockedBlockWAL)
	defer env.finalize()

	nodeID := nodeIDs[0]
	oldBlocks := env.bf.GetBlocks(batchSize, 1)
	env.sendBlocksToNode(nodeID, 0*time.Second, oldBlocks...)
	require.True(env.t, env.ensureStoreContainsBlocksNoWait(nodeID, oldBlocks))

	newBlocks := env.bf.GetBlocksFrom(branchSize, 1, state.OriginL1Commitment(), 2)
	env.sendBlocksToNode(nodeID, 0*time.Second, newBlocks...)
	require.True(env.t, env.ensureStoreContainsBlocksNoWait(nodeID, newBlocks))

	oldCommitment := oldBlocks[len(oldBlocks)-1].L1Commitment()
	newCommitment := newBlocks[len(newBlocks)-1].L1Commitment()
	require.True(env.t, env.sendAndEnsureCompletedChainFetchStateDiff(oldCommitment, newCommitment, oldBlocks, newBlocks, nodeID, 1, 0*time.Second))
}

// Single node network. Checks if block cache is cleaned via state manager
// timer events.
func TestBlockCacheCleaningAuto(t *testing.T) {
	nodeIDs := gpa.MakeTestNodeIDs(1)
	smTimers := NewStateManagerTimers()
	smTimers.BlockCacheBlocksInCacheDuration = 300 * time.Millisecond
	smTimers.BlockCacheBlockCleaningPeriod = 70 * time.Millisecond
	env := newTestEnv(t, nodeIDs, smGPAUtils.NewMockedBlockWAL, smTimers)
	defer env.finalize()

	nodeID := nodeIDs[0]
	blocks := env.bf.GetBlocks(6, 2)

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
