package smGPAUtils

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
	origState := factory.GetState(blocks[numberOfBlocks-1].L1Commitment())
	fileName := "TestWriteReadDifferentStores.snap"
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o666)
	require.NoError(t, err)
	err = writeStateToFile(origState, fileName, f)
	require.NoError(t, err)
	mutations, err := readStateFromFile(fileName)
	require.NoError(t, err)
	err = os.Remove(fileName)
	require.NoError(t, err)

	store := state.NewStore(mapdb.NewMapDB())
	stateDraft := store.NewOriginStateDraft()
	// stateDraft, err := store.NewEmptyStateDraft(blocks[numberOfBlocks-1].PreviousL1Commitment())
	// require.NoError(t, err)
	mutations.ApplyTo(stateDraft)
	block := store.Commit(stateDraft)
	// Commitments should be equal too, but currently only trie roots are equal.
	// hashes are not equal, because previous l1 commitment of resulting block is
	// nil, as origin state draft was used to commit it.
	// require.True(t, blocks[numberOfBlocks-1].L1Commitment().Equals(block.L1Commitment()))
	require.True(t, blocks[numberOfBlocks-1].TrieRoot().Equals(block.TrieRoot()))

	resultState, err := store.StateByTrieRoot(blocks[numberOfBlocks-1].TrieRoot())
	require.NoError(t, err)
	require.True(t, origState.TrieRoot().Equals(resultState.TrieRoot()))
	require.Equal(t, origState.BlockIndex(), resultState.BlockIndex())
	require.Equal(t, origState.Timestamp(), resultState.Timestamp())
	require.True(t, origState.PreviousL1Commitment().Equals(resultState.PreviousL1Commitment()))

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
	iterateFun(resultState, func(entry *commonEntry, value []byte) { entry.valueResult = value })

	for _, entry := range commonState {
		require.Equal(t, entry.valueOrig, entry.valueResult)
	}
}
