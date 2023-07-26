package sm_gpa

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain/statemanager/sm_gpa/sm_gpa_utils"
	"github.com/iotaledger/wasp/packages/chain/statemanager/sm_gpa/sm_inputs"
	"github.com/iotaledger/wasp/packages/chain/statemanager/sm_snapshots"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/origin"
	"github.com/iotaledger/wasp/packages/state"
)

var newEmptySnapshotManagerFun = func(_, _ state.Store, _ sm_gpa_utils.TimeProvider, _ *logger.Logger) sm_snapshots.SnapshotManager {
	return sm_snapshots.NewEmptySnapshotManager()
}

// Single node network. 8 blocks are sent to state manager. The result is checked
// by checking the store and sending consensus requests, which force the access
// of the blocks.
func TestBasic(t *testing.T) {
	nodeIDs := gpa.MakeTestNodeIDs(1)
	env := newTestEnv(t, nodeIDs, sm_gpa_utils.NewMockedTestBlockWAL, newEmptySnapshotManagerFun)
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
	smParameters := NewStateManagerParameters()
	smParameters.StateManagerGetBlockRetry = 100 * time.Millisecond
	env := newTestEnv(t, nodeIDs, sm_gpa_utils.NewMockedTestBlockWAL, newEmptySnapshotManagerFun, smParameters)
	defer env.finalize()

	blocks := env.bf.GetBlocks(16, 1)
	env.sendBlocksToNode(nodeIDs[0], smParameters.StateManagerGetBlockRetry, blocks...)
	require.True(env.t, env.ensureStoreContainsBlocksNoWait(nodeIDs[0], blocks))

	// Nodes are checked sequentially
	lastCommitment := blocks[7].L1Commitment()
	for i := 1; i < len(nodeIDs); i++ {
		env.t.Logf("Sequential: waiting for blocks ending with %s to be available on node %s...", lastCommitment, nodeIDs[i].ShortString())
		require.True(env.t, env.sendAndEnsureCompletedConsensusStateProposal(lastCommitment, nodeIDs[i], 100, smParameters.StateManagerGetBlockRetry))
		require.True(env.t, env.sendAndEnsureCompletedConsensusDecidedState(lastCommitment, nodeIDs[i], 8, smParameters.StateManagerGetBlockRetry))
		require.True(env.t, env.ensureStoreContainsBlocksNoWait(nodeIDs[i], blocks[:8]))
	}
	// Nodes are checked in parallel
	lastCommitment = blocks[15].L1Commitment()
	lastOutput := env.bf.GetAliasOutput(lastCommitment)
	cspInputs := make(map[gpa.NodeID]gpa.Input)
	cspRespChans := make(map[gpa.NodeID]<-chan interface{})
	for i := 1; i < len(nodeIDs); i++ {
		nodeID := nodeIDs[i]
		cspInputs[nodeID], cspRespChans[nodeID] = sm_inputs.NewConsensusStateProposal(context.Background(), lastOutput)
	}
	env.tc.WithInputs(cspInputs).RunAll()
	for nodeID, cspRespChan := range cspRespChans {
		env.t.Logf("Parallel: waiting for blocks ending with %s to be available on node %s...", lastCommitment, nodeID.ShortString())
		require.True(env.t, env.ensureCompletedConsensusStateProposal(cspRespChan, 10, smParameters.StateManagerGetBlockRetry))
	}
	cdsInputs := make(map[gpa.NodeID]gpa.Input)
	cdsRespChans := make(map[gpa.NodeID]<-chan state.State)
	for i := 1; i < len(nodeIDs); i++ {
		nodeID := nodeIDs[i]
		cdsInputs[nodeID], cdsRespChans[nodeID] = sm_inputs.NewConsensusDecidedState(context.Background(), lastOutput)
	}
	env.tc.WithInputs(cdsInputs).RunAll()
	for nodeID, cdsRespChan := range cdsRespChans {
		env.t.Logf("Parallel: waiting for state %s on node %s", lastCommitment, nodeID.ShortString())
		require.True(env.t, env.ensureCompletedConsensusDecidedState(cdsRespChan, lastCommitment, 16, smParameters.StateManagerGetBlockRetry))
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
	smParameters := NewStateManagerParameters()
	smParameters.StateManagerGetBlockRetry = 100 * time.Millisecond
	env := newTestEnv(t, nodeIDs, sm_gpa_utils.NewMockedTestBlockWAL, newEmptySnapshotManagerFun, smParameters)
	defer env.finalize()

	lastCommitment := origin.L1Commitment(nil, 0)

	testIterationFun := func(i int, baseCommitment *state.L1Commitment, incrementFactor ...uint64) []state.Block {
		env.t.Logf("Iteration %v: generating %v blocks and sending them to nodes", i, iterationSize)
		blocks := env.bf.GetBlocksFrom(iterationSize, 1, baseCommitment, incrementFactor...)
		env.sendBlocksToRandomNode(nodeIDs, smParameters.StateManagerGetBlockRetry, blocks...)
		for _, nodeID := range nodeIDs {
			randCommitment := blocks[rand.Intn(iterationSize-1)].L1Commitment() // Do not pick the last state/blocks
			t.Logf("Iteration %v: sending ConsensusDecidedState for commitment %s to node %s",
				i, randCommitment, nodeID.ShortString())
			require.True(env.t, env.sendAndEnsureCompletedConsensusDecidedState(randCommitment, nodeID, maxRetriesPerIteration, smParameters.StateManagerGetBlockRetry))
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
		require.True(env.t, env.sendAndEnsureCompletedConsensusDecidedState(lastCommitment, nodeID, maxRetriesPerIteration, smParameters.StateManagerGetBlockRetry))
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
		require.True(env.t, env.sendAndEnsureCompletedConsensusDecidedState(lastCommitment, nodeID, maxRetriesPerIteration, smParameters.StateManagerGetBlockRetry))
		require.True(env.t, env.ensureStoreContainsBlocksNoWait(nodeID, newBlocks))
	}
	newCommitment := lastCommitment

	// ChainFetchStateDiff request
	for _, nodeID := range env.nodeIDs {
		env.t.Logf("Requesting state for Mempool from node %s", nodeID.ShortString())
		require.True(env.t, env.sendAndEnsureCompletedChainFetchStateDiff(oldCommitment, newCommitment, oldBlocks, newBlocks, nodeID, maxRetriesPerIteration, smParameters.StateManagerGetBlockRetry))
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
	smParameters := NewStateManagerParameters()
	smParameters.StateManagerGetBlockRetry = 100 * time.Millisecond
	env := newTestEnv(t, nodeIDs, sm_gpa_utils.NewMockedTestBlockWAL, newEmptySnapshotManagerFun, smParameters)
	defer env.finalize()

	mainBlocks := env.bf.GetBlocks(mainSize, 1)
	branchIndex := randomFrom + rand.Intn(randomTo-randomFrom)
	branchBlocks := env.bf.GetBlocksFrom(branchSize, 1, mainBlocks[branchIndex].L1Commitment(), 2)

	env.sendBlocksToNode(nodeIDs[0], smParameters.StateManagerGetBlockRetry, mainBlocks...)
	require.True(env.t, env.ensureStoreContainsBlocksNoWait(nodeIDs[0], mainBlocks))
	env.sendBlocksToNode(nodeIDs[0], smParameters.StateManagerGetBlockRetry, branchBlocks...)
	require.True(env.t, env.ensureStoreContainsBlocksNoWait(nodeIDs[0], branchBlocks))

	oldCommitment := mainBlocks[len(mainBlocks)-1].L1Commitment()
	newCommitment := branchBlocks[len(branchBlocks)-1].L1Commitment()
	oldBlocks := mainBlocks[branchIndex+1:]
	for _, nodeID := range nodeIDs[1:] {
		env.t.Logf("Sending ChainFetchStateDiff for old commitment %s and new commitment %s to node %s", oldCommitment, newCommitment, nodeID.ShortString())
		require.True(env.t, env.sendAndEnsureCompletedChainFetchStateDiff(oldCommitment, newCommitment, oldBlocks, branchBlocks, nodeID, maxRetries, smParameters.StateManagerGetBlockRetry))
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
	env := newTestEnv(t, nodeIDs, sm_gpa_utils.NewMockedTestBlockWAL, newEmptySnapshotManagerFun)
	defer env.finalize()

	nodeID := nodeIDs[0]
	blocks := env.bf.GetBlocks(1, 1)
	env.sendBlocksToNode(nodeID, 0*time.Second, blocks[0])
	require.True(env.t, env.ensureStoreContainsBlocksNoWait(nodeID, blocks))

	oldCommitment := origin.L1Commitment(nil, 0)
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
	env := newTestEnv(t, nodeIDs, sm_gpa_utils.NewMockedTestBlockWAL, newEmptySnapshotManagerFun)
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
	env := newTestEnv(t, nodeIDs, sm_gpa_utils.NewMockedTestBlockWAL, newEmptySnapshotManagerFun)
	defer env.finalize()

	nodeID := nodeIDs[0]
	oldBlocks := env.bf.GetBlocks(batchSize, 1)
	env.sendBlocksToNode(nodeID, 0*time.Second, oldBlocks...)
	require.True(env.t, env.ensureStoreContainsBlocksNoWait(nodeID, oldBlocks))

	newBlocks := env.bf.GetBlocksFrom(branchSize, 1, origin.L1Commitment(nil, 0), 2)
	env.sendBlocksToNode(nodeID, 0*time.Second, newBlocks...)
	require.True(env.t, env.ensureStoreContainsBlocksNoWait(nodeID, newBlocks))

	oldCommitment := oldBlocks[len(oldBlocks)-1].L1Commitment()
	newCommitment := newBlocks[len(newBlocks)-1].L1Commitment()
	require.True(env.t, env.sendAndEnsureCompletedChainFetchStateDiff(oldCommitment, newCommitment, oldBlocks, newBlocks, nodeID, 1, 0*time.Second))
}

// Two node setting.
//  1. A batch of 10 consecutive blocks is generated, each of them is sent
//     to the first node.
//  2. A batch of 5 consecutive blocks is branched from block 4. Each of
//     the blocks is sent to the first node.
//  3. Second node is configured to download snapshot at index 7 of both branches
//  4. A ChainFetchStateDiff request is sent for the branch as a new and
//     and original batch as old.
func TestMempoolSnapshotInTheMiddle(t *testing.T) {
	batchSize := 10
	branchSize := 5
	branchIndex := 4
	snapshottedIndex := 7

	nodeIDs := gpa.MakeTestNodeIDs(2)
	newMockedSnapshotManagerFun := func(origStore, nodeStore state.Store, timeProvider sm_gpa_utils.TimeProvider, log *logger.Logger) sm_snapshots.SnapshotManager {
		return sm_snapshots.NewMockedSnapshotManager(t, 0, origStore, nodeStore, 0*time.Second, 0*time.Second, timeProvider, log)
	}
	smParameters := NewStateManagerParameters()
	smParameters.StateManagerGetBlockRetry = 100 * time.Millisecond
	env := newTestEnv(t, nodeIDs, sm_gpa_utils.NewMockedTestBlockWAL, newMockedSnapshotManagerFun, smParameters)
	defer env.finalize()

	oldBlocks := env.bf.GetBlocks(batchSize, 1)
	newBlocks := env.bf.GetBlocksFrom(branchSize, 1, oldBlocks[branchIndex].L1Commitment(), 2)
	oldSnapshottedBlock := oldBlocks[snapshottedIndex]
	newSnapshottedBlock := newBlocks[snapshottedIndex-branchIndex-1]
	snapm1 := env.snapms[nodeIDs[1]].(*sm_snapshots.MockedSnapshotManager)
	snapm1.SnapshotReady(sm_snapshots.NewSnapshotInfo(oldSnapshottedBlock.StateIndex(), oldSnapshottedBlock.L1Commitment()))
	snapm1.SnapshotReady(sm_snapshots.NewSnapshotInfo(newSnapshottedBlock.StateIndex(), newSnapshottedBlock.L1Commitment()))
	snapm1.UpdateAsync()

	env.sendBlocksToNode(nodeIDs[0], 0*time.Second, oldBlocks...)
	require.True(env.t, env.ensureStoreContainsBlocksNoWait(nodeIDs[0], oldBlocks))

	env.sendBlocksToNode(nodeIDs[0], 0*time.Second, newBlocks...)
	require.True(env.t, env.ensureStoreContainsBlocksNoWait(nodeIDs[0], newBlocks))

	oldCommitment := oldBlocks[len(oldBlocks)-1].L1Commitment()
	newCommitment := newBlocks[len(newBlocks)-1].L1Commitment()
	responseCh := env.sendChainFetchStateDiff(oldCommitment, newCommitment, nodeIDs[1])
	require.True(env.t, snapm1.WaitSnapshotLoadRequestCount(1, 10*time.Millisecond, 100)) // To allow snapshot manager to receive load old state snapshot request
	env.sendTimerTickToNodes(100 * time.Millisecond)                                      // To check the response from snapshot manager about loaded old state snapshot; timer tick is not necessary: any input would be suitable
	require.True(env.t, snapm1.WaitSnapshotLoadedCount(1, 10*time.Millisecond, 100))      // To allow snapshot manager thread to wake up and respond
	require.True(env.t, snapm1.WaitSnapshotLoadRequestCount(2, 10*time.Millisecond, 100)) // To allow snapshot manager to receive load new state snapshot request
	env.sendTimerTickToNodes(100 * time.Millisecond)                                      // To check the response from snapshot manager about loaded new state snapshot; timer tick is not necessary: any input would be suitable
	require.True(env.t, snapm1.WaitSnapshotLoadedCount(2, 10*time.Millisecond, 100))      // To allow snapshot manager thread to wake up and respond
	require.True(env.t, env.ensureCompletedChainFetchStateDiff(responseCh, oldBlocks[branchIndex+1:], newBlocks, 10, 100*time.Millisecond))
}

// Single node setting, pruning leaves 10 historic blocks.
//   - 11 blocks are added into the store one by one; each time it is checked if
//     all of the added blocks are in the store (none of them got pruned).
//   - 9 blocks are added into the store one by one; each time it is checked if
//     only the newest block and 10 others are still in store and the remaining
//     blocks are pruned.
func TestPruningSequentially(t *testing.T) {
	blocksToKeep := 10
	blockCount := 20

	nodeIDs := gpa.MakeTestNodeIDs(1)
	nodeID := nodeIDs[0]
	smParameters := NewStateManagerParameters()
	smParameters.PruningMinStatesToKeep = blocksToKeep
	env := newTestEnv(t, nodeIDs, sm_gpa_utils.NewEmptyTestBlockWAL, newEmptySnapshotManagerFun, smParameters)
	defer env.finalize()

	blocks := env.bf.GetBlocks(blockCount, 1)
	for i := 0; i <= blocksToKeep; i++ {
		env.sendBlocksToNode(nodeID, 0*time.Second, blocks[i])
		require.True(env.t, env.ensureStoreContainsBlocksNoWait(nodeID, blocks[:i+1]))
		for j := 0; j <= i; j++ {
			env.checkBlock(nodeID, blocks[j])
		}
	}
	for i := blocksToKeep + 1; i < blockCount; i++ {
		lastExistingBlockIndex := i - blocksToKeep
		env.sendBlocksToNode(nodeID, 0*time.Second, blocks[i])
		require.True(env.t, env.ensureStoreContainsBlocksNoWait(nodeID, blocks[lastExistingBlockIndex:i+1]))
		for j := 0; j < lastExistingBlockIndex; j++ {
			env.doesNotContainBlock(nodeID, blocks[j])
		}
		for j := lastExistingBlockIndex; j <= i; j++ {
			env.checkBlock(nodeID, blocks[j])
		}
	}
}

// Single node setting
//   - pruning leaves 10000 historic blocks.
//   - 20 blocks are committed, none of them are pruned.
//   - state manager is configured to leave 10 historic blocks after pruning.
//   - another block is committed to trigger pruning, 11 blocks (origin+10 committed
//     blocks) are pruned.
func TestPruningMany(t *testing.T) {
	blocksToKeep := 10
	blocksToSend := 20

	nodeIDs := gpa.MakeTestNodeIDs(1)
	nodeID := nodeIDs[0]
	smParameters := NewStateManagerParameters()
	smParameters.PruningMinStatesToKeep = blocksToKeep // Also initializes chain with this value in governance contract
	env := newTestEnv(t, nodeIDs, sm_gpa_utils.NewEmptyTestBlockWAL, newEmptySnapshotManagerFun, smParameters)
	defer env.finalize()

	sm, ok := env.sms[nodeID]
	require.True(env.t, ok)
	sm.(*stateManagerGPA).parameters.PruningMinStatesToKeep = 10000

	blocks := env.bf.GetBlocks(blocksToSend+1, 1)
	env.sendBlocksToNode(nodeID, 0*time.Second, blocks[:blocksToSend]...)
	require.True(env.t, env.ensureStoreContainsBlocksNoWait(nodeID, blocks[:blocksToSend]))

	sm.(*stateManagerGPA).parameters.PruningMinStatesToKeep = blocksToKeep
	env.sendBlocksToNode(nodeID, 0*time.Second, blocks[blocksToSend])
	lastExistingBlockIndex := blocksToSend - blocksToKeep
	require.True(env.t, env.ensureStoreContainsBlocksNoWait(nodeID, blocks[lastExistingBlockIndex:]))
	for j := 0; j < lastExistingBlockIndex; j++ {
		env.doesNotContainBlock(nodeID, blocks[j])
	}
	for j := lastExistingBlockIndex; j <= blocksToSend; j++ {
		env.checkBlock(nodeID, blocks[j])
	}
}

// Single node setting
//   - pruning leaves 10000 historic blocks.
//   - 30 blocks are committed, none of them are pruned.
//   - state manager is configured to leave 10 historic blocks after pruning but
//     not to delete more than 8 blocks in one pruning run.
//   - a block is committed several times to trigger pruning, each time pruning
//     8 blocks or (in the case of last iteration) as many as needed (but not more
//     than 8) to leave 10 historic blocks (in addition to recently committed block).
func TestPruningTooMuch(t *testing.T) {
	blocksToKeep := 10
	blocksToSend := 30
	blocksToPrune := 8

	nodeIDs := gpa.MakeTestNodeIDs(1)
	nodeID := nodeIDs[0]
	smParameters := NewStateManagerParameters()
	smParameters.PruningMinStatesToKeep = blocksToKeep // Also initializes chain with this value in governance contract
	smParameters.PruningMaxStatesToDelete = blocksToPrune
	env := newTestEnv(t, nodeIDs, sm_gpa_utils.NewEmptyTestBlockWAL, newEmptySnapshotManagerFun, smParameters)
	defer env.finalize()

	sm, ok := env.sms[nodeID]
	require.True(env.t, ok)
	sm.(*stateManagerGPA).parameters.PruningMinStatesToKeep = 10000

	blocks := env.bf.GetBlocks(blocksToSend, 1)
	env.sendBlocksToNode(nodeID, 0*time.Second, blocks...)
	require.True(env.t, env.ensureStoreContainsBlocksNoWait(nodeID, blocks))

	sm.(*stateManagerGPA).parameters.PruningMinStatesToKeep = blocksToKeep
	lastExistingBlockIndex := -1 // Origin block is not in blocks array
	lastExistingBlockIndexExpected := blocksToSend - blocksToKeep - 1
	for lastExistingBlockIndex < lastExistingBlockIndexExpected {
		newBlocks := env.bf.GetBlocks(1, 1)
		env.sendBlocksToNode(nodeID, 0*time.Second, newBlocks[0])
		require.True(env.t, env.ensureStoreContainsBlocksNoWait(nodeID, newBlocks))
		blocks = append(blocks, newBlocks[0])
		lastExistingBlockIndexExpected++
		lastExistingBlockIndex += blocksToPrune
		if lastExistingBlockIndex > lastExistingBlockIndexExpected {
			lastExistingBlockIndex = lastExistingBlockIndexExpected
		}
		for j := 0; j < lastExistingBlockIndex; j++ {
			env.doesNotContainBlock(nodeID, blocks[j])
		}
		for j := lastExistingBlockIndex; j < len(blocks); j++ {
			env.checkBlock(nodeID, blocks[j])
		}
	}
}

// Two nodes setting: first node is making snapshots, the other is using them.
//   - 30 blocks are committed to the first node.
//   - Update is triggered (in both nodes) and created snapshots are available
//     to be used by state manager.
//   - The other node is requested to obtain snapshotted state, and several not
//     snapshotted states. The results are checked.
func TestSnapshots(t *testing.T) {
	blockCount := 30
	snapshotCreatePeriod := uint32(5)
	snapshotCreateTime := 1 * time.Second
	snapshotLoadTime := 2 * time.Second
	snapshotCount := uint32(blockCount) / snapshotCreatePeriod
	snapshotCreatedFun := func(index uint32) bool {
		return (index+1)%snapshotCreatePeriod == 0
	}
	requestedStateIndex1 := uint32(5)
	requestedStateIndex2 := uint32(14)
	requestedStateIndex3 := uint32(23)
	timerTickPeriod := 150 * time.Millisecond

	nodeIDs := gpa.MakeTestNodeIDs(2)
	nodeIDFirst := nodeIDs[0]
	nodeIDOther := nodeIDs[1]
	newEmptyTestBlockWALFun := func(gpa.NodeID) sm_gpa_utils.TestBlockWAL { return sm_gpa_utils.NewEmptyTestBlockWAL() }
	newMockedSnapshotManagerFun := func(nodeID gpa.NodeID, origStore, nodeStore state.Store, tp sm_gpa_utils.TimeProvider, log *logger.Logger) sm_snapshots.SnapshotManager {
		if nodeID.Equals(nodeIDFirst) {
			return sm_snapshots.NewMockedSnapshotManager(t, snapshotCreatePeriod, origStore, nodeStore, snapshotCreateTime, snapshotLoadTime, tp, log)
		}
		return sm_snapshots.NewMockedSnapshotManager(t, 0, origStore, nodeStore, snapshotCreateTime, snapshotLoadTime, tp, log)
	}
	smParameters := NewStateManagerParameters()
	smParameters.SnapshotManagerUpdatePeriod = 2 * time.Second
	env := newVariedTestEnv(t, nodeIDs, newEmptyTestBlockWALFun, newMockedSnapshotManagerFun, smParameters)
	snapMFirst := env.snapms[nodeIDFirst].(*sm_snapshots.MockedSnapshotManager)
	snapMOther := env.snapms[nodeIDOther].(*sm_snapshots.MockedSnapshotManager)
	snapMFirst.SetAfterSnapshotCreated(func(snapshotInfo sm_snapshots.SnapshotInfo) {
		<-env.timeProvider.After(100 * time.Millisecond) // Other node knows about the snapshot a bit later
		snapMOther.SnapshotReady(snapshotInfo)
	})
	defer env.finalize()

	blocks := env.bf.GetBlocks(blockCount, 1)
	snapshotInfos := make([]sm_snapshots.SnapshotInfo, len(blocks))
	for i := range snapshotInfos {
		snapshotInfos[i] = sm_snapshots.NewSnapshotInfo(blocks[i].StateIndex(), blocks[i].L1Commitment())
	}
	type expectedValues struct {
		snapshotReady  bool // Ready to be picked up by snapshot manager's Update
		snapshotExists bool // Already picked up by snapshot manager's Update; available to node
		blockCommitted bool
	}
	checkBlocksInNodeFun := func(expected []expectedValues, nodeID gpa.NodeID) {
		snapM, ok := env.snapms[nodeID]
		require.True(env.t, ok)
		store, ok := env.stores[nodeID]
		require.True(env.t, ok)
		for i := range expected {
			env.t.Logf("Checking snapshot/block index %v at node %v", snapshotInfos[i].StateIndex(), nodeID)
			require.Equal(env.t, expected[i].snapshotReady, snapM.(*sm_snapshots.MockedSnapshotManager).IsSnapshotReady(snapshotInfos[i]))
			require.Equal(env.t, expected[i].snapshotExists, snapM.SnapshotExists(snapshotInfos[i].StateIndex(), snapshotInfos[i].Commitment()))
			require.Equal(env.t, expected[i].blockCommitted, store.HasTrieRoot(snapshotInfos[i].TrieRoot()))
		}
	}
	expectedFirst := make([]expectedValues, len(blocks))
	expectedOther := make([]expectedValues, len(blocks))
	checkBlocksFun := func() {
		checkBlocksInNodeFun(expectedFirst, nodeIDFirst)
		checkBlocksInNodeFun(expectedOther, nodeIDOther)
	}
	checkBlocksFun()                                                                 // At start no blocks/snapshots are in any node
	env.sendTimerTickToNodes(0 * time.Second)                                        // Initial timer tick to send first snapshot manager Update request
	require.True(env.t, snapMFirst.WaitNodeUpdateCount(1, 10*time.Millisecond, 100)) // Time for first snapshot manager Update request to propagate to snapshot manager (and do nothing)
	require.True(env.t, snapMOther.WaitNodeUpdateCount(1, 10*time.Millisecond, 100)) // Time for first snapshot manager Update request to propagate to snapshot manager (and do nothing)

	// Blocks are sent to the first node: they are committed there, snapshots are being produced, but not yet available
	env.sendBlocksToNode(nodeIDFirst, 0*time.Second, blocks...)
	require.True(env.t, snapMFirst.WaitSnapshotCreateRequestCount(snapshotCount, 10*time.Millisecond, 100))
	for i := range blocks {
		expectedFirst[i].blockCommitted = true
	}
	checkBlocksFun()

	// Time is passing, snapshots are produced and are ready in the first node; Update hasn't picked them up yet
	for i := 0; i < 7; i++ {
		env.sendTimerTickToNodes(timerTickPeriod) // Timer tick is not necessary; it's just a way to advance artificial timer
	}
	require.True(env.t, snapMFirst.WaitSnapshotCreatedCount(snapshotCount, 10*time.Millisecond, 100)) // To allow threads, that "create snapshots", to wake up
	for i := range blocks {
		if snapshotCreatedFun(uint32(i)) {
			expectedFirst[i].snapshotReady = true
		}
	}
	checkBlocksFun()

	// Some more time passes, produced snapshots are visible in other node too; Update hasn't picked them up yet
	env.sendTimerTickToNodes(timerTickPeriod)                                                                 // Timer tick is not necessary; it's just a way to advance artificial timer
	require.True(env.t, snapMFirst.WaitSnapshotCreateFinalisedCount(snapshotCount, 10*time.Millisecond, 100)) // To allow threads, that "create snapshots", to wake up
	for i := range blocks {
		expectedOther[i].snapshotReady = expectedFirst[i].snapshotReady
	}
	checkBlocksFun()

	// More time passes, Update event is triggered in both nodes, snapshots are available for state managers of both nodes
	for i := 0; i < 7; i++ {
		env.sendTimerTickToNodes(timerTickPeriod) // Only the last timer tick is necessary as it sends Update request to snapshot manager
	}
	require.True(env.t, snapMFirst.WaitNodeUpdateCount(2, 10*time.Millisecond, 100)) // Time for snapshot manager Update request to propagate to snapshot manager
	require.True(env.t, snapMOther.WaitNodeUpdateCount(2, 10*time.Millisecond, 100)) // Time for snapshot manager Update request to propagate to snapshot manager
	for i := range blocks {
		expectedFirst[i].snapshotExists = expectedFirst[i].snapshotReady
		expectedOther[i].snapshotExists = expectedOther[i].snapshotReady
	}
	checkBlocksFun()

	loadRequestsCount := uint32(0)
	sendAndEnsureCompletedConsensusStateProposalWithWaitFun := func(snapshotInfo sm_snapshots.SnapshotInfo) {
		respCh := env.sendConsensusStateProposal(snapshotInfo.Commitment(), nodeIDOther)
		loadRequestsCount++
		require.True(env.t, snapMOther.WaitSnapshotLoadRequestCount(loadRequestsCount, 10*time.Millisecond, 100)) // Time for load snapshot request to propagate to snapshot manager
		for i := 0; i < 14; i++ {
			env.sendTimerTickToNodes(timerTickPeriod) // Timer tick is not necessary; it's just a way to advance artificial timer
		}
		require.True(env.t, snapMOther.WaitSnapshotLoadedCount(loadRequestsCount, 10*time.Millisecond, 100)) // To allow snapshot manager thread to wake up and respond
		require.True(env.t, env.ensureCompletedConsensusStateProposal(respCh, 2, timerTickPeriod))
	}

	// Request for other node to have state, which contains snapshot; snapshot is downloaded, no other blocks are committed
	require.True(env.t, snapshotCreatedFun(requestedStateIndex2))
	sendAndEnsureCompletedConsensusStateProposalWithWaitFun(snapshotInfos[requestedStateIndex2])
	expectedOther[requestedStateIndex2].blockCommitted = true
	checkBlocksFun()

	// Request for other node to have state, which is one index after snapshot; snapshot and the requested block are downloaded
	require.True(env.t, snapshotCreatedFun(requestedStateIndex1-1))
	sendAndEnsureCompletedConsensusStateProposalWithWaitFun(snapshotInfos[requestedStateIndex1])
	expectedOther[requestedStateIndex1].blockCommitted = true
	expectedOther[requestedStateIndex1-1].blockCommitted = true
	checkBlocksFun()

	// Request for other node to have state, which is one index before snapshot; all the blocks up to previous snapshot and the previous snapshot are downloaded
	require.True(env.t, snapshotCreatedFun(requestedStateIndex3+1))
	sendAndEnsureCompletedConsensusStateProposalWithWaitFun(snapshotInfos[requestedStateIndex3])
	for i := requestedStateIndex3; i > requestedStateIndex3-snapshotCreatePeriod; i-- {
		expectedOther[i].blockCommitted = true
	}
	checkBlocksFun()
}

// Single node network. Checks if block cache is cleaned via state manager
// timer events.
func TestBlockCacheCleaningAuto(t *testing.T) {
	nodeIDs := gpa.MakeTestNodeIDs(1)
	smParameters := NewStateManagerParameters()
	smParameters.BlockCacheBlocksInCacheDuration = 300 * time.Millisecond
	smParameters.BlockCacheBlockCleaningPeriod = 70 * time.Millisecond
	env := newTestEnv(t, nodeIDs, sm_gpa_utils.NewEmptyTestBlockWAL, newEmptySnapshotManagerFun, smParameters)
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
