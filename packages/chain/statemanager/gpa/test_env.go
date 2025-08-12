package gpa

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"fortio.org/safecast"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/packages/chain/statemanager/gpa/inputs"
	gpautils "github.com/iotaledger/wasp/v2/packages/chain/statemanager/gpa/utils"
	"github.com/iotaledger/wasp/v2/packages/chain/statemanager/snapshots"
	"github.com/iotaledger/wasp/v2/packages/chain/statemanager/utils"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/kvstore/mapdb"
	"github.com/iotaledger/wasp/v2/packages/metrics"
	"github.com/iotaledger/wasp/v2/packages/origin"
	"github.com/iotaledger/wasp/v2/packages/parameters/parameterstest"
	"github.com/iotaledger/wasp/v2/packages/state"
	"github.com/iotaledger/wasp/v2/packages/state/statetest"
	"github.com/iotaledger/wasp/v2/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/v2/packages/util"
	"github.com/iotaledger/wasp/v2/packages/util/timeutil"
	"github.com/iotaledger/wasp/v2/packages/vm/core/migrations/allmigrations"
)

type testEnv struct {
	t          *testing.T
	bf         *gpautils.BlockFactory
	nodeIDs    []gpa.NodeID
	parameters StateManagerParameters
	sms        map[gpa.NodeID]gpa.GPA
	stores     map[gpa.NodeID]state.Store
	snapms     map[gpa.NodeID]snapshots.SnapshotManager
	tc         *gpa.TestContext
	log        log.Logger
}

func newTestEnv(
	t *testing.T,
	nodeIDs []gpa.NodeID,
	createWALFun func() gpautils.TestBlockWAL,
	createSnapMFun func(origStore, nodeStore state.Store, tp timeutil.TimeProvider, log log.Logger) snapshots.SnapshotManager,
	parametersOpt ...StateManagerParameters,
) *testEnv {
	result := newTestEnvNoNodes(t, parametersOpt...)
	result.addNodes(nodeIDs, createWALFun, createSnapMFun)
	return result
}

// Commented to please the linter; left for completeness
/*func newTestEnvVariedNodes(
	t *testing.T,
	nodeIDs []gpa.NodeID,
	createWALFun func(gpa.NodeID) utils.TestBlockWAL,
	createSnapMFun func(nodeID gpa.NodeID, origStore, nodeStore state.Store, tp utils.TimeProvider, log log.Logger) snapshots.SnapshotManager,
	parametersOpt ...StateManagerParameters,
) *testEnv {
	result := newTestEnvNoNodes(t, parametersOpt...)
	result.addVariedNodes(nodeIDs, createWALFun, createSnapMFun)
	return result
}*/

func newTestEnvNoNodes(
	t *testing.T,
	parametersOpt ...StateManagerParameters,
) *testEnv {
	var bf *gpautils.BlockFactory
	var parameters StateManagerParameters
	var chainInitParameters gpautils.BlockFactoryCallArguments
	if len(parametersOpt) > 0 {
		parameters = parametersOpt[0]
		int32Val, err := safecast.Convert[int32](parameters.PruningMinStatesToKeep)
		if err != nil {
			panic(fmt.Sprintf("integer overflow in PruningMinStatesToKeep: %v overflows int32", parameters.PruningMinStatesToKeep))
		}
		chainInitParameters = gpautils.BlockFactoryCallArguments{BlockKeepAmount: int32Val}
		bf = gpautils.NewBlockFactory(t, chainInitParameters)
	} else {
		parameters = NewStateManagerParameters()
		bf = gpautils.NewBlockFactory(t)
	}

	log := testlogger.NewLogger(t)
	parameters.TimeProvider = timeutil.NewArtificialTimeProvider()
	return &testEnv{
		t:          t,
		bf:         bf,
		parameters: parameters,
		log:        log,
	}
}

func (teT *testEnv) addNodes(
	nodeIDs []gpa.NodeID,
	createWALFun func() gpautils.TestBlockWAL,
	createSnapMFun func(origStore, nodeStore state.Store, tp timeutil.TimeProvider, log log.Logger) snapshots.SnapshotManager,
) {
	createWALVariedFun := func(gpa.NodeID) gpautils.TestBlockWAL {
		return createWALFun()
	}
	createSnapMVariedFun := func(nodeID gpa.NodeID, origStore, nodeStore state.Store, tp timeutil.TimeProvider, log log.Logger) snapshots.SnapshotManager {
		return createSnapMFun(origStore, nodeStore, tp, log)
	}
	teT.addVariedNodes(nodeIDs, createWALVariedFun, createSnapMVariedFun)
}

func (teT *testEnv) addVariedNodes(
	nodeIDs []gpa.NodeID,
	createWALFun func(gpa.NodeID) gpautils.TestBlockWAL,
	createSnapMFun func(nodeID gpa.NodeID, origStore, nodeStore state.Store, tp timeutil.TimeProvider, log log.Logger) snapshots.SnapshotManager,
) {
	sms := make(map[gpa.NodeID]gpa.GPA)
	stores := make(map[gpa.NodeID]state.Store)
	snapms := make(map[gpa.NodeID]snapshots.SnapshotManager)
	for _, nodeID := range nodeIDs {
		var err error
		smLog := teT.log.NewChildLogger(nodeID.ShortString())
		nr := utils.NewNodeRandomiser(nodeID, nodeIDs, smLog)
		wal := createWALFun(nodeID)
		store := statetest.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB())
		snapshotManager := createSnapMFun(nodeID, teT.bf.GetStore(), store, teT.parameters.TimeProvider, smLog)
		loadedSnapshotStateIndex := snapshotManager.GetLoadedSnapshotStateIndex()
		stores[nodeID] = store
		sms[nodeID], err = New(loadedSnapshotStateIndex, nr, wal, store, mockStateManagerMetrics(), smLog, teT.parameters)
		require.NoError(teT.t, err)
		snapms[nodeID] = snapshotManager
		origin.InitChain(allmigrations.LatestSchemaVersion, store, teT.bf.GetChainInitParameters(), iotago.ObjectID{}, 0, parameterstest.L1Mock)
	}
	teT.nodeIDs = nodeIDs
	teT.sms = sms
	teT.snapms = snapms
	teT.stores = stores
	teT.tc = gpa.NewTestContext(sms).WithOutputHandler(func(nodeID gpa.NodeID, outputOrig gpa.Output) {
		output, ok := outputOrig.(StateManagerOutput)
		require.True(teT.t, ok)
		snapshotManager, ok := teT.snapms[nodeID]
		require.True(teT.t, ok)
		for _, snapshotInfo := range output.TakeBlocksCommitted() {
			snapshotManager.BlockCommittedAsync(snapshotInfo)
		}
		for _, nextInput := range output.TakeNextInputs() {
			teT.tc.WithInputs(map[gpa.NodeID]gpa.Input{nodeID: nextInput}).RunAll()
		}
	})
}

func (teT *testEnv) finalize() {
}

func (teT *testEnv) checkBlock(nodeID gpa.NodeID, origBlock state.Block) {
	store, ok := teT.stores[nodeID]
	require.True(teT.t, ok)
	gpautils.CheckBlockInStore(teT.t, store, origBlock)
}

func (teT *testEnv) doesNotContainBlock(nodeID gpa.NodeID, block state.Block) {
	store, ok := teT.stores[nodeID]
	require.True(teT.t, ok)
	require.False(teT.t, store.HasTrieRoot(block.TrieRoot()))
}

func (teT *testEnv) sendBlocksToNode(nodeID gpa.NodeID, timeStep time.Duration, blocks ...state.Block) {
	// If `ConsensusBlockProduced` is sent to the node, the node has definitely obtained all the blocks
	// needed to commit this block. This is ensured by consensus.
	require.True(teT.t, teT.sendAndEnsureCompletedConsensusStateProposal(blocks[0].PreviousL1Commitment(), nodeID, 100, timeStep))
	for i := range blocks {
		teT.t.Logf("Supplying block %s to node %s", blocks[i].L1Commitment(), nodeID.ShortString())
		teT.sendAndEnsureCompletedConsensusBlockProduced(blocks[i], nodeID, 100, timeStep)
	}

	store, ok := teT.stores[nodeID]
	require.True(teT.t, ok)
	err := store.SetLatest(blocks[len(blocks)-1].TrieRoot())
	require.NoError(teT.t, err)
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

func (teT *testEnv) sendConsensusBlockProduced(block state.Block, nodeID gpa.NodeID) <-chan state.Block {
	input, responseCh := inputs.NewConsensusBlockProduced(context.Background(), teT.bf.GetStateDraft(block))
	teT.tc.WithInputs(map[gpa.NodeID]gpa.Input{nodeID: input}).RunAll()
	return responseCh
}

func (teT *testEnv) ensureCompletedConsensusBlockProduced(respChan <-chan state.Block, maxTimeIterations int, timeStep time.Duration) bool {
	return teT.ensureTrue("response from ConsensusBlockProduced", func() bool {
		select {
		case block := <-respChan:
			require.NotNil(teT.t, block)
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

func (teT *testEnv) sendConsensusStateProposal(commitment *state.L1Commitment, nodeID gpa.NodeID) <-chan any {
	input, responseCh := inputs.NewConsensusStateProposal(context.Background(), teT.bf.GetAnchor(commitment))
	teT.tc.WithInputs(map[gpa.NodeID]gpa.Input{nodeID: input}).RunAll()
	return responseCh
}

func (teT *testEnv) ensureCompletedConsensusStateProposal(respChan <-chan any, maxTimeIterations int, timeStep time.Duration) bool {
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
	input, responseCh := inputs.NewConsensusDecidedState(context.Background(), teT.bf.GetAnchor(commitment))
	teT.tc.WithInputs(map[gpa.NodeID]gpa.Input{nodeID: input}).RunAll()
	return responseCh
}

func (teT *testEnv) ensureCompletedConsensusDecidedState(respChan <-chan state.State, expectedCommitment *state.L1Commitment, maxTimeIterations int, timeStep time.Duration) bool {
	return teT.ensureTrue("response from ConsensusDecidedState", func() bool {
		select {
		case s := <-respChan:
			gpautils.CheckStateInStore(teT.t, teT.bf.GetStore(), s)
			require.True(teT.t, expectedCommitment.TrieRoot().Equals(s.TrieRoot()))
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

func (teT *testEnv) sendChainFetchStateDiff(oldCommitment, newCommitment *state.L1Commitment, nodeID gpa.NodeID) <-chan *inputs.ChainFetchStateDiffResults {
	input, responseCh := inputs.NewChainFetchStateDiff(context.Background(), teT.bf.GetAnchor(oldCommitment), teT.bf.GetAnchor(newCommitment))
	teT.tc.WithInputs(map[gpa.NodeID]gpa.Input{nodeID: input}).RunAll()
	return responseCh
}

func (teT *testEnv) ensureCompletedChainFetchStateDiff(respChan <-chan *inputs.ChainFetchStateDiffResults, expectedOldBlocks, expectedNewBlocks []state.Block, maxTimeIterations int, timeStep time.Duration) bool {
	return teT.ensureTrue("response from ChainFetchStateDiff", func() bool {
		select {
		case cfsdr := <-respChan:
			newStateTrieRoot := cfsdr.GetNewState().TrieRoot()
			lastNewBlockTrieRoot := expectedNewBlocks[len(expectedNewBlocks)-1].TrieRoot()
			teT.t.Logf("Checking trie roots: expected %s, obtained %s", lastNewBlockTrieRoot, newStateTrieRoot)
			require.True(teT.t, newStateTrieRoot.Equals(lastNewBlockTrieRoot))
			gpautils.CheckStateInStore(teT.t, teT.bf.GetStore(), cfsdr.GetNewState())
			requireEqualsFun := func(expected, received []state.Block) {
				teT.t.Logf("\tExpected %v elements, obtained %v elements", len(expected), len(received))
				require.Equal(teT.t, len(expected), len(received))
				for i := range expected {
					teT.t.Logf("\tchecking %v-th element: expected %s, received %s", i, expected[i].L1Commitment(), received[i].L1Commitment())
					require.True(teT.t, expected[i].Equals(received[i]))
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

func (teT *testEnv) ensureStoreContainsBlocksNoWait(nodeID gpa.NodeID, blocks []state.Block) bool {
	return teT.ensureTrue("store to contain blocks", func() bool {
		for _, block := range blocks {
			commitment := block.L1Commitment()
			teT.t.Logf("Checking block %s on node %s...", commitment, nodeID.ShortString())
			store, ok := teT.stores[nodeID]
			require.True(teT.t, ok)
			if store.HasTrieRoot(commitment.TrieRoot()) {
				teT.t.Logf("Node %s contains block %s", nodeID.ShortString(), commitment)
			} else {
				teT.t.Logf("Node %s does not contain block %s", nodeID.ShortString(), commitment)
				return false
			}
		}
		return true
	}, 1, 0*time.Second)
}

// --------

func (teT *testEnv) ensureTrue(title string, predicate func() bool, maxTimeIterations int, timeStep time.Duration) bool {
	if predicate() {
		return true
	}
	for i := 1; i < maxTimeIterations; i++ {
		teT.t.Logf("Waiting for %s iteration %v", title, i)
		teT.sendTimerTickToNodes(timeStep)
		if predicate() {
			return true
		}
	}
	return false
}

func (teT *testEnv) sendTimerTickToNodes(delay time.Duration) {
	now := teT.parameters.TimeProvider.GetNow().Add(delay)
	teT.parameters.TimeProvider.SetNow(now)
	teT.t.Logf("Time %v is sent to nodes %s", now, util.SliceShortString(teT.nodeIDs))
	teT.sendInputToNodes(func(_ gpa.NodeID) gpa.Input {
		return inputs.NewStateManagerTimerTick(now)
	})
}

func (teT *testEnv) sendInputToNodes(makeInputFun func(gpa.NodeID) gpa.Input) {
	inputs := make(map[gpa.NodeID]gpa.Input)
	for _, nodeID := range teT.nodeIDs {
		inputs[nodeID] = makeInputFun(nodeID)
	}
	teT.tc.WithInputs(inputs).RunAll()
}

func mockStateManagerMetrics() *metrics.ChainStateManagerMetrics {
	return metrics.NewChainMetricsProvider().GetChainMetrics(isc.EmptyChainID()).StateManager
}
