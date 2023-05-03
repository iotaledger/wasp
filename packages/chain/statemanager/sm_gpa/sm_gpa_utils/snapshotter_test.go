package sm_gpa_utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

func TestWriteReadDifferentStores(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()

	var err error
	numberOfBlocks := 10
	factory := NewBlockFactory(t)
	blocks := factory.GetBlocks(numberOfBlocks, 1)
	lastCommitment := blocks[numberOfBlocks-1].L1Commitment()
	origState := factory.GetState(lastCommitment)
	origSnapshot := factory.GetSnapshot(lastCommitment)
	fileName := "TestWriteReadDifferentStores.snap"
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o666)
	require.NoError(t, err)
	err = writeSnapshotToFile(lastCommitment.TrieRoot(), origSnapshot, fileName, f)
	require.NoError(t, err)
	newTrieRoot, newSnapshot, err := readSnapshotFromFile(fileName)
	require.NoError(t, err)
	require.True(t, lastCommitment.TrieRoot().Equals(newTrieRoot))
	err = os.Remove(fileName)
	require.NoError(t, err)

	store := state.NewStore(mapdb.NewMapDB())
	err = store.RestoreSnapshot(newTrieRoot, newSnapshot)
	require.NoError(t, err)

	newBlock, err := store.BlockByTrieRoot(newTrieRoot)
	require.NoError(t, err)
	require.True(t, lastCommitment.TrieRoot().Equals(newBlock.TrieRoot()))
	require.True(t, lastCommitment.BlockHash().Equals(newBlock.Hash()))

	newState, err := store.StateByTrieRoot(newTrieRoot)
	require.NoError(t, err)
	require.True(t, origState.TrieRoot().Equals(newState.TrieRoot()))
	require.Equal(t, origState.BlockIndex(), newState.BlockIndex())
	require.Equal(t, origState.Timestamp(), newState.Timestamp())
	require.True(t, origState.PreviousL1Commitment().Equals(newState.PreviousL1Commitment()))

	type commonEntry struct {
		valueOrig   []byte
		valueResult []byte
	}
	commonState := make(map[kv.Key]*commonEntry)
	iterateFun := func(iterState state.State, setValueFun func(*commonEntry, []byte)) {
		iterState.Iterate(kv.EmptyPrefix, func(key kv.Key, value []byte) bool {
			entry, ok := commonState[key]
			if !ok {
				entry = &commonEntry{}
				commonState[key] = entry
			}
			setValueFun(entry, value)
			return true
		})
	}
	iterateFun(origState, func(entry *commonEntry, value []byte) { entry.valueOrig = value })
	iterateFun(newState, func(entry *commonEntry, value []byte) { entry.valueResult = value })

	for _, entry := range commonState {
		require.Equal(t, entry.valueOrig, entry.valueResult)
	}
}
