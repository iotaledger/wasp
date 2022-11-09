package smGPA

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/kvstore/mapdb"
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/trie.go/trie"
	"github.com/iotaledger/wasp/packages/chain/aaa2/cons/gr"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smGPAUtils"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smInputs"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smUtils"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

// Single node network. 8 blocks are sent to state manager. The result is checked
// by sending consensus requests, which force the access of the blocks.
func TestBasic(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()

	chainID, blocks, stateOutputs := smGPAUtils.GetBlocks(t, 8, 1)
	nodeID := gpa.MakeTestNodeIDs("Node", 1)[0]
	_, sm := createStateManagerGpa(t, chainID, nodeID, []gpa.NodeID{nodeID}, smGPAUtils.NewMockedBlockWAL(), log)
	tc := gpa.NewTestContext(map[gpa.NodeID]gpa.GPA{nodeID: sm})
	sendBlocksToNode(t, tc, nodeID, blocks)

	cspInput, cspRespChan := smInputs.NewConsensusStateProposal(context.Background(), stateOutputs[7])
	tc.WithInputs(map[gpa.NodeID]gpa.Input{nodeID: cspInput}).RunAll()
	require.NoError(t, requireReceiveAnything(cspRespChan, 5*time.Second))
	commitment, err := state.L1CommitmentFromBytes(stateOutputs[7].GetAliasOutput().StateMetadata)
	require.NoError(t, err)
	cdsInput, cdsRespChan := smInputs.NewConsensusDecidedState(context.Background(), stateOutputs[7])
	tc.WithInputs(map[gpa.NodeID]gpa.Input{nodeID: cdsInput}).RunAll()
	require.NoError(t, requireReceiveVState(t, cdsRespChan, 8, &commitment, 5*time.Second))
}

// 10 nodes in a network. 8 blocks are sent to state manager of the first node.
// The result is checked by sending consensus requests to all the other 9 nodes,
// which force the access (and retrieval) of the blocks. For successful retrieval,
// several timer events are required for nodes to try to request blocks from peers.
func TestManyNodes(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()

	smTimers := NewStateManagerTimers()
	smTimers.StateManagerGetBlockRetry = 100 * time.Millisecond

	chainID, blocks, stateOutputs := smGPAUtils.GetBlocks(t, 16, 1)
	nodeIDs := gpa.MakeTestNodeIDs("Node", 10)
	sms := make(map[gpa.NodeID]gpa.GPA)
	for _, nodeID := range nodeIDs {
		_, sm := createStateManagerGpa(t, chainID, nodeID, nodeIDs, smGPAUtils.NewMockedBlockWAL(), log, smTimers)
		sms[nodeID] = sm
	}
	tc := gpa.NewTestContext(sms)
	sendBlocksToNode(t, tc, nodeIDs[0], blocks)

	//Nodes are checked sequentially
	var result bool
	now := time.Now()
	for i := 1; i < len(nodeIDs); i++ {
		cspInput, cspRespChan := smInputs.NewConsensusStateProposal(context.Background(), stateOutputs[7])
		tc.WithInputs(map[gpa.NodeID]gpa.Input{nodeIDs[i]: cspInput}).RunAll()
		t.Logf("Sequential: waiting for blocks ending with %s to be available on node %s...", blocks[7].GetHash(), nodeIDs[i])
		now, result = requireReceiveAnythingNTimes(t, tc, cspRespChan, 10, nodeIDs, now, 200*time.Millisecond)
		require.True(t, result)
		commitment, err := state.L1CommitmentFromBytes(stateOutputs[7].GetAliasOutput().StateMetadata)
		require.NoError(t, err)
		cdsInput, cdsRespChan := smInputs.NewConsensusDecidedState(context.Background(), stateOutputs[7])
		tc.WithInputs(map[gpa.NodeID]gpa.Input{nodeIDs[i]: cdsInput}).RunAll()
		require.NoError(t, requireReceiveVState(t, cdsRespChan, 8, &commitment, 5*time.Second))
	}
	//Nodes are checked in parallel
	cspInputs := make(map[gpa.NodeID]gpa.Input)
	cspRespChans := make(map[gpa.NodeID]<-chan interface{})
	for i := 1; i < len(nodeIDs); i++ {
		nodeID := nodeIDs[i]
		cspInputs[nodeID], cspRespChans[nodeID] = smInputs.NewConsensusStateProposal(context.Background(), stateOutputs[15])
	}
	tc.WithInputs(cspInputs).RunAll()
	for nodeID, cspRespChan := range cspRespChans {
		t.Logf("Parallel: waiting for blocks ending with %s to be available on node %s...", blocks[15].GetHash(), nodeID)
		now, result = requireReceiveAnythingNTimes(t, tc, cspRespChan, 10, nodeIDs, now, 200*time.Millisecond)
		require.True(t, result)
	}
	commitment, err := state.L1CommitmentFromBytes(stateOutputs[15].GetAliasOutput().StateMetadata)
	require.NoError(t, err)
	cdsInputs := make(map[gpa.NodeID]gpa.Input)
	cdsRespChans := make(map[gpa.NodeID]<-chan *consGR.StateMgrDecidedState)
	for i := 1; i < len(nodeIDs); i++ {
		nodeID := nodeIDs[i]
		cdsInputs[nodeID], cdsRespChans[nodeID] = smInputs.NewConsensusDecidedState(context.Background(), stateOutputs[15])
	}
	tc.WithInputs(cdsInputs).RunAll()
	for nodeID, cdsRespChan := range cdsRespChans {
		t.Logf("Parallel: waiting for state %s on node %s", commitment, nodeID)
		require.NoError(t, requireReceiveVState(t, cdsRespChan, 16, &commitment, 5*time.Second))
	}
}

// Single node network. Checks if block cache is cleaned via state manager
// timer events.
func TestBlockCacheCleaningAuto(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()

	tp := smGPAUtils.NewArtifficialTimeProvider()
	smTimers := NewStateManagerTimers(tp)
	smTimers.BlockCacheBlocksInCacheDuration = 300 * time.Millisecond
	smTimers.BlockCacheBlockCleaningPeriod = 70 * time.Millisecond

	chainID, blocks, _ := smGPAUtils.GetBlocks(t, 6, 2)
	nodeID := gpa.MakeTestNodeIDs("Node", 1)[0]
	_, sm := createStateManagerGpa(t, chainID, nodeID, []gpa.NodeID{nodeID}, smGPAUtils.NewEmptyBlockWAL(), log, smTimers)
	tc := gpa.NewTestContext(map[gpa.NodeID]gpa.GPA{nodeID: sm})

	advanceTimeAndTimerTickFun := func(advance time.Duration) {
		tp.SetNow(tp.GetNow().Add(advance))
		tc.WithInputs(map[gpa.NodeID]gpa.Input{nodeID: smInputs.NewStateManagerTimerTick(tp.GetNow())}).RunAll()
	}

	blockCache := sm.(*stateManagerGPA).blockCache
	blockCache.AddBlock(blocks[0])
	blockCache.AddBlock(blocks[1])
	require.NotNil(t, blockCache.GetBlock(1, blocks[0].GetHash()))
	require.NotNil(t, blockCache.GetBlock(2, blocks[1].GetHash()))
	advanceTimeAndTimerTickFun(100 * time.Millisecond)
	require.NotNil(t, blockCache.GetBlock(1, blocks[0].GetHash()))
	require.NotNil(t, blockCache.GetBlock(2, blocks[1].GetHash()))
	blockCache.AddBlock(blocks[2])
	require.NotNil(t, blockCache.GetBlock(1, blocks[0].GetHash()))
	require.NotNil(t, blockCache.GetBlock(2, blocks[1].GetHash()))
	require.NotNil(t, blockCache.GetBlock(2, blocks[2].GetHash()))
	advanceTimeAndTimerTickFun(100 * time.Millisecond)
	require.NotNil(t, blockCache.GetBlock(1, blocks[0].GetHash()))
	require.NotNil(t, blockCache.GetBlock(2, blocks[1].GetHash()))
	require.NotNil(t, blockCache.GetBlock(2, blocks[2].GetHash()))
	blockCache.AddBlock(blocks[3])
	require.NotNil(t, blockCache.GetBlock(1, blocks[0].GetHash()))
	require.NotNil(t, blockCache.GetBlock(2, blocks[1].GetHash()))
	require.NotNil(t, blockCache.GetBlock(2, blocks[2].GetHash()))
	require.NotNil(t, blockCache.GetBlock(3, blocks[3].GetHash()))
	advanceTimeAndTimerTickFun(80 * time.Millisecond)
	tc.WithInputs(map[gpa.NodeID]gpa.Input{nodeID: smInputs.NewStateManagerTimerTick(tp.GetNow())}).RunAll()
	require.NotNil(t, blockCache.GetBlock(1, blocks[0].GetHash()))
	require.NotNil(t, blockCache.GetBlock(2, blocks[1].GetHash()))
	require.NotNil(t, blockCache.GetBlock(2, blocks[2].GetHash()))
	require.NotNil(t, blockCache.GetBlock(3, blocks[3].GetHash()))
	advanceTimeAndTimerTickFun(100 * time.Millisecond)
	blockCache.AddBlock(blocks[4])
	require.Nil(t, blockCache.GetBlock(1, blocks[0].GetHash()))
	require.Nil(t, blockCache.GetBlock(2, blocks[1].GetHash()))
	require.NotNil(t, blockCache.GetBlock(2, blocks[2].GetHash()))
	require.NotNil(t, blockCache.GetBlock(3, blocks[3].GetHash()))
	require.NotNil(t, blockCache.GetBlock(3, blocks[4].GetHash()))
	advanceTimeAndTimerTickFun(100 * time.Millisecond)
	require.Nil(t, blockCache.GetBlock(2, blocks[2].GetHash()))
	require.NotNil(t, blockCache.GetBlock(3, blocks[3].GetHash()))
	require.NotNil(t, blockCache.GetBlock(3, blocks[4].GetHash()))
	advanceTimeAndTimerTickFun(100 * time.Millisecond)
	require.Nil(t, blockCache.GetBlock(3, blocks[3].GetHash()))
	require.NotNil(t, blockCache.GetBlock(3, blocks[4].GetHash()))
	advanceTimeAndTimerTickFun(200 * time.Millisecond)
	require.Nil(t, blockCache.GetBlock(3, blocks[4].GetHash()))
}

func createStateManagerGpa(t *testing.T, chainID *isc.ChainID, me gpa.NodeID, nodeIDs []gpa.NodeID, wal smGPAUtils.BlockWAL, log *logger.Logger, timers ...StateManagerTimers) (smUtils.NodeRandomiser, gpa.GPA) {
	log = log.Named(me.String()).Named("c-" + chainID.ShortString())
	nr := smUtils.NewNodeRandomiser(me, nodeIDs, log)
	store := mapdb.NewMapDB()
	sm, err := New(chainID, nr, wal, store, log, timers...)
	require.NoError(t, err)
	return nr, sm
}

func requireReceiveNoError(t *testing.T, errChan <-chan (error), timeout time.Duration) error { //nolint:gocritic
	select {
	case err := <-errChan:
		require.NoError(t, err)
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("Waiting to receive no error timeouted")
	}
}

func requireReceiveAnything(anyChan <-chan (interface{}), timeout time.Duration) error { //nolint:gocritic
	select {
	case <-anyChan:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("Waiting to receive anything timeouted")
	}
}

func requireReceiveVState(t *testing.T, respChan <-chan (*consGR.StateMgrDecidedState), index uint32, l1c *state.L1Commitment, timeout time.Duration) error { //nolint:gocritic
	select {
	case smds := <-respChan:
		require.Equal(t, smds.VirtualStateAccess.BlockIndex(), index)
		require.True(t, smds.StateBaseline.IsValid())
		require.True(t, state.EqualCommitments(trie.RootCommitment(smds.VirtualStateAccess.TrieNodeStore()), l1c.StateCommitment))
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("Waiting to receive state timeouted")
	}
}

func sendBlocksToNode(t *testing.T, tc *gpa.TestContext, nodeID gpa.NodeID, blocks []state.Block) {
	for i := range blocks {
		cbpInput, cbpRespChan := smInputs.NewChainBlockProduced(context.Background(), blocks[i])
		t.Logf("Supplying block %s to node %s", blocks[i].GetHash(), nodeID)
		tc.WithInputs(map[gpa.NodeID]gpa.Input{nodeID: cbpInput}).RunAll()
		require.NoError(t, requireReceiveNoError(t, cbpRespChan, 5*time.Second))
	}
}

func sendTimerTickToNodes(t *testing.T, tc *gpa.TestContext, nodeIDs []gpa.NodeID, now time.Time) {
	t.Logf("Time %v is sent to nodes %v", now, nodeIDs)
	inputs := make(map[gpa.NodeID]gpa.Input)
	for i := range nodeIDs {
		inputs[nodeIDs[i]] = smInputs.NewStateManagerTimerTick(now)
	}
	tc.WithInputs(inputs).RunAll()
}

func requireReceiveAnythingNTimes(
	t *testing.T, tc *gpa.TestContext, anyChan <-chan (interface{}), n int, //nolint:gocritic
	nodeIDs []gpa.NodeID, now time.Time, delay time.Duration,
) (time.Time, bool) {
	newNow := now
	for j := 0; j < n; j++ {
		t.Logf("\t...iteration %v", j)
		if requireReceiveAnything(anyChan, 0*time.Second) == nil {
			return newNow, true
		}
		newNow = newNow.Add(delay)
		sendTimerTickToNodes(t, tc, nodeIDs, newNow)
	}
	return newNow, false
}
