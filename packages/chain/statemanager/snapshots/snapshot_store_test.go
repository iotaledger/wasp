package snapshots

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/chain/statemanager/sm_gpa/sm_gpa_utils"
	"github.com/iotaledger/wasp/packages/state"
)

func TestNewerSnapshotKeepsOlderSnapshot(t *testing.T) {
	twoSnapshotsCheckEnds(t, func(t *testing.T, _storeOrig, storeNew state.Store, intermediateSnapshot, lastSnapshot *bytes.Buffer, blocks []state.Block) {
		intermediateTrieRoot := blocks[0].TrieRoot()
		lastTrieRoot := blocks[len(blocks)-1].TrieRoot()

		err := storeNew.RestoreSnapshot(intermediateTrieRoot, intermediateSnapshot)
		require.NoError(t, err)
		require.True(t, storeNew.HasTrieRoot(intermediateTrieRoot))

		err = storeNew.RestoreSnapshot(lastTrieRoot, lastSnapshot)
		require.NoError(t, err)
		require.True(t, storeNew.HasTrieRoot(intermediateTrieRoot))
		require.True(t, storeNew.HasTrieRoot(lastTrieRoot))
	})
}

func TestOlderSnapshotKeepsNewerSnapshot(t *testing.T) {
	twoSnapshotsCheckEnds(t, func(t *testing.T, _storeOrig, storeNew state.Store, intermediateSnapshot, lastSnapshot *bytes.Buffer, blocks []state.Block) {
		intermediateTrieRoot := blocks[0].TrieRoot()
		lastTrieRoot := blocks[len(blocks)-1].TrieRoot()

		err := storeNew.RestoreSnapshot(lastTrieRoot, lastSnapshot)
		require.NoError(t, err)
		require.True(t, storeNew.HasTrieRoot(lastTrieRoot))

		err = storeNew.RestoreSnapshot(intermediateTrieRoot, intermediateSnapshot)
		require.NoError(t, err)
		require.True(t, storeNew.HasTrieRoot(intermediateTrieRoot))
		require.True(t, storeNew.HasTrieRoot(lastTrieRoot))
	})
}

func TestFillTheBlocksBetweenSnapshots(t *testing.T) {
	twoSnapshotsCheckEnds(t, func(t *testing.T, storeOrig, storeNew state.Store, intermediateSnapshot, lastSnapshot *bytes.Buffer, blocks []state.Block) {
		intermediateTrieRoot := blocks[0].TrieRoot()
		lastTrieRoot := blocks[len(blocks)-1].TrieRoot()
		err := storeNew.RestoreSnapshot(lastTrieRoot, lastSnapshot)
		require.NoError(t, err)
		err = storeNew.RestoreSnapshot(intermediateTrieRoot, intermediateSnapshot)
		require.NoError(t, err)
		require.True(t, storeNew.HasTrieRoot(intermediateTrieRoot))
		require.True(t, storeNew.HasTrieRoot(lastTrieRoot))
		for i := 1; i < len(blocks); i++ {
			stateDraft, err := storeNew.NewEmptyStateDraft(blocks[i].PreviousL1Commitment())
			require.NoError(t, err)
			blocks[i].Mutations().ApplyTo(stateDraft)
			block := storeNew.Commit(stateDraft)
			require.True(t, blocks[i].TrieRoot().Equals(block.TrieRoot()))
			require.True(t, blocks[i].Hash().Equals(block.Hash()))
		}
		for i := 1; i < len(blocks)-1; i++ { // blocks[i] and blocsk[len(blocks)-1] will be checked in `twoSnapshotsCheckEnds`
			sm_gpa_utils.CheckBlockInStore(t, storeNew, blocks[i])
			sm_gpa_utils.CheckStateInStores(t, storeOrig, storeNew, blocks[i].L1Commitment())
		}
	})
}

func twoSnapshotsCheckEnds(t *testing.T, performTestFun func(t *testing.T, storeOrig, storeNew state.Store, intermediateSnapshot, lastSnapshot *bytes.Buffer, blocks []state.Block)) {
	numberOfBlocks := 10
	intermediateBlockIndex := 4

	factory := sm_gpa_utils.NewBlockFactory(t)
	blocks := factory.GetBlocks(numberOfBlocks, 1)
	storeOrig := factory.GetStore()
	storeNew := state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB())

	intermediateBlock := blocks[intermediateBlockIndex]
	intermediateCommitment := intermediateBlock.L1Commitment()
	intermediateSnapshot := new(bytes.Buffer)
	err := storeOrig.TakeSnapshot(intermediateCommitment.TrieRoot(), intermediateSnapshot)
	require.NoError(t, err)

	lastBlock := blocks[len(blocks)-1]
	lastCommitment := lastBlock.L1Commitment()
	lastSnapshot := new(bytes.Buffer)
	err = storeOrig.TakeSnapshot(lastCommitment.TrieRoot(), lastSnapshot)
	require.NoError(t, err)

	performTestFun(t, storeOrig, storeNew, intermediateSnapshot, lastSnapshot, blocks[intermediateBlockIndex:])

	sm_gpa_utils.CheckBlockInStore(t, storeNew, intermediateBlock)
	sm_gpa_utils.CheckStateInStores(t, storeOrig, storeNew, intermediateCommitment)
	sm_gpa_utils.CheckBlockInStore(t, storeNew, lastBlock)
	sm_gpa_utils.CheckStateInStores(t, storeOrig, storeNew, lastCommitment)
}
