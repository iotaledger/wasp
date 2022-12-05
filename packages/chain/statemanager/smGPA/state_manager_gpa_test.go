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
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
)

// Single node network. 8 blocks are sent to state manager. The result is checked
// by sending consensus requests, which force the access of the blocks.
func TestBasic(t *testing.T) {
	nodeIDs := gpa.MakeTestNodeIDs("Node", 1)
	env := newTestEnv(t, nodeIDs, smGPAUtils.NewMockedBlockWAL)
	defer env.finalize()

	nodeID := nodeIDs[0]
	blocks, stateOutputs := env.bf.GetBlocks(8, 1)
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

	blocks, stateOutputs := env.bf.GetBlocks(16, 1)
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
//  1. This is repeated 3 times, resulting in 3 consecutive batches of blocks:
//     1.1 Chain of 10 blocks are generated; each of them are sent to a random node
//     1.2 A randomly chosen block in a batch is chosen and is approved
//     1.3 A successful change of state of state manager is waited for; each check
//     fires a timer event to force the exchange of blocks between nodes.
//  2. The last block of the last batch is approved and 1.3 is repeated.
//  3. A random block is chosen in the second batch and 1.1, 1.2, 1.3 and 2 are
//     repeated branching from this block
//  4. A common ancestor (mempool) request is sent to state manager for first and
//     second batch ends.
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

	lastAliasOutput := env.bf.GetOriginOutput()
	lastCommitment := state.OriginL1Commitment()

	confirmAndWaitFun := func(stateOutput *isc.AliasOutputWithID) bool {
		env.sendInputToNodes(func(_ gpa.NodeID) gpa.Input {
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

	testIterationFun := func(i int, baseAliasOutput *isc.AliasOutputWithID, baseCommitment *state.L1Commitment, incrementFactor ...uint64) ([]*isc.AliasOutputWithID, []state.Block) {
		env.t.Logf("Iteration %v: generating %v blocks and sending them to nodes", i, iterationSize)
		blocks, aliasOutputs := env.bf.GetBlocksFrom(iterationSize, 1, baseCommitment, baseAliasOutput, incrementFactor...)
		for _, block := range blocks {
			env.sendBlocksToNode(nodeIDs[rand.Intn(nodeCount)], block)
		}
		confirmOutput := aliasOutputs[rand.Intn(iterationSize-1)] // Do not confirm the last state/blocks
		t.Logf("Iteration %v: approving output index %v, id %v", i, confirmOutput.GetStateIndex(), confirmOutput)
		require.True(t, confirmAndWaitFun(confirmOutput))
		return aliasOutputs, blocks
	}

	var branchAliasOutput *isc.AliasOutputWithID
	var branchCommitment *state.L1Commitment
	oldBlocks := make([]state.Block, 0)
	for i := 0; i < iterationCount; i++ {
		aliasOutputs, blocks := testIterationFun(i, lastAliasOutput, lastCommitment, 1)
		lastAliasOutput = aliasOutputs[iterationSize-1]
		lastCommitment = blocks[iterationSize-1].L1Commitment()
		if i == 1 {
			branchIndex := rand.Intn(iterationSize)
			branchAliasOutput = aliasOutputs[branchIndex]
			branchCommitment = blocks[branchIndex].L1Commitment()
			oldBlocks = append(oldBlocks, blocks[branchIndex+1:]...)
		} else if i == 2 {
			oldBlocks = append(oldBlocks, blocks...)
		}
	}
	require.True(t, confirmAndWaitFun(lastAliasOutput))
	oldAliasOutput := lastAliasOutput

	// Branching from the middle of the second iteration
	require.NotNil(t, branchAliasOutput)
	lastAliasOutput = branchAliasOutput
	lastCommitment = branchCommitment
	newBlocks := make([]state.Block, 0)
	for i := 0; i < iterationCount; i++ {
		aliasOutputs, blocks := testIterationFun(i, lastAliasOutput, lastCommitment, 2)
		lastAliasOutput = aliasOutputs[iterationSize-1]
		lastCommitment = blocks[iterationSize-1].L1Commitment()
		newBlocks = append(newBlocks, blocks...)
	}
	require.True(t, confirmAndWaitFun(lastAliasOutput))
	newAliasOutput := lastAliasOutput

	// Common ancestor request
	for _, nodeID := range env.nodeIDs {
		request, responseCh := smInputs.NewMempoolStateRequest(context.Background(), oldAliasOutput, newAliasOutput)
		env.t.Logf("Requesting state for Mempool from node %s", nodeID)
		env.tc.WithInputs(map[gpa.NodeID]gpa.Input{nodeID: request}).RunAll()
		require.NoError(env.t, env.requireReceiveMempoolResults(responseCh, oldBlocks, newBlocks, 5*time.Second))
	}
}

// 15 nodes setting.
//  1. A batch of 20 consecutive blocks is generated, each of them is sent to the
//     first node.
//  2. A block is selected at random in the middle (from index 5 to 12), and a
//     branch of 5 nodes is generated.
//  3. For each node a common ancestor (mempool) request is sent and the successful
//     completion is waited for; each check fires a timer event to force
//     the exchange of blocks between nodes.
func TestMempoolRequest(t *testing.T) {
	nodeCount := 15
	mainSize := 20
	randomFrom := 5
	randomTo := 12
	branchSize := 5
	maxRetries := 100

	nodeIDs := gpa.MakeTestNodeIDs("Node", nodeCount)
	smTimers := NewStateManagerTimers()
	smTimers.StateManagerGetBlockRetry = 100 * time.Millisecond
	env := newTestEnv(t, nodeIDs, smGPAUtils.NewMockedBlockWAL, smTimers)
	defer env.finalize()

	mainBlocks, mainAliasOutputs := env.bf.GetBlocks(mainSize, 1)
	branchIndex := randomFrom + rand.Intn(randomTo-randomFrom)
	branchBlocks, branchAliasOutputs := env.bf.GetBlocksFrom(branchSize, 1, mainBlocks[branchIndex].L1Commitment(), mainAliasOutputs[branchIndex], 2)

	env.sendBlocksToNode(nodeIDs[0], mainBlocks...)
	env.sendBlocksToNode(nodeIDs[0], branchBlocks...)

	respChans := make(map[gpa.NodeID]<-chan *smInputs.MempoolStateRequestResults)
	oldAliasOutput := mainAliasOutputs[len(mainAliasOutputs)-1]
	newAliasOutput := branchAliasOutputs[len(branchAliasOutputs)-1]
	env.sendInputToNodes(func(nodeID gpa.NodeID) gpa.Input {
		input, respChan := smInputs.NewMempoolStateRequest(context.Background(), oldAliasOutput, newAliasOutput)
		respChans[nodeID] = respChan
		return input
	})

	oldBlocks := mainBlocks[branchIndex+1:]
	for _, nodeID := range nodeIDs[1:] {
		env.t.Logf("Waiting for response from node %s", nodeID)
		respChan, ok := respChans[nodeID]
		require.True(env.t, ok)
		received := false
		for i := 0; i < maxRetries && !received; i++ {
			t.Logf("\twaiting for blocks to propagate through nodes: %v", i)
			if env.requireReceiveMempoolResults(respChan, oldBlocks, branchBlocks, 0*time.Second) == nil {
				received = true
			} else {
				env.sendTimerTickToNodes(smTimers.StateManagerGetBlockRetry)
			}
		}
		require.True(env.t, received)
	}
}

// Single node setting.
//  1. A single block is generated and sent to the node.
//  2. A common ancestor (mempool) request is sent for block 1 as a new block
//     and block 0 as an old block.
func TestMempoolRequestFirstStep(t *testing.T) {
	nodeIDs := gpa.MakeTestNodeIDs("Node", 1)
	env := newTestEnv(t, nodeIDs, smGPAUtils.NewMockedBlockWAL)
	defer env.finalize()

	nodeID := nodeIDs[0]
	blocks, stateOutputs := env.bf.GetBlocks(1, 1)
	env.sendBlocksToNode(nodeID, blocks[0])

	var respChan <-chan *smInputs.MempoolStateRequestResults
	oldAliasOutput := env.bf.GetOriginOutput()
	newAliasOutput := stateOutputs[0]
	env.sendInputToNodes(func(nodeID gpa.NodeID) gpa.Input {
		var input gpa.Input
		input, respChan = smInputs.NewMempoolStateRequest(context.Background(), oldAliasOutput, newAliasOutput)
		return input
	})
	oldBlocks := make([]state.Block, 0)
	err := env.requireReceiveMempoolResults(respChan, oldBlocks, blocks, 0*time.Second)
	require.NoError(env.t, err)
}

// Single node setting.
//  1. A batch of 10 consecutive blocks is generated, each of them is sent to the node.
//  2. A common ancestor (mempool) request is sent for block 10 as a new block
//     and block 5 as an old block.
func TestMempoolRequestNoBranch(t *testing.T) {
	batchSize := 10
	middleBlock := 4

	nodeIDs := gpa.MakeTestNodeIDs("Node", 1)
	env := newTestEnv(t, nodeIDs, smGPAUtils.NewMockedBlockWAL)
	defer env.finalize()

	nodeID := nodeIDs[0]
	blocks, stateOutputs := env.bf.GetBlocks(batchSize, 1)
	env.sendBlocksToNode(nodeID, blocks...)

	var respChan <-chan *smInputs.MempoolStateRequestResults
	oldAliasOutput := stateOutputs[middleBlock]
	newAliasOutput := stateOutputs[len(stateOutputs)-1]
	env.sendInputToNodes(func(nodeID gpa.NodeID) gpa.Input {
		var input gpa.Input
		input, respChan = smInputs.NewMempoolStateRequest(context.Background(), oldAliasOutput, newAliasOutput)
		return input
	})
	oldBlocks := make([]state.Block, 0)
	newBlocks := blocks[middleBlock+1:]
	err := env.requireReceiveMempoolResults(respChan, oldBlocks, newBlocks, 0*time.Second)
	require.NoError(env.t, err)
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
	blocks, _ := env.bf.GetBlocks(6, 2)

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
