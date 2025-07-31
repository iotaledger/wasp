package state_test

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/kvstore"
	"github.com/iotaledger/wasp/v2/packages/kvstore/mapdb"
	"github.com/iotaledger/wasp/v2/packages/state"
	"github.com/iotaledger/wasp/v2/packages/state/statetest"
)

type stateSM struct {
	store state.Store
	draft state.StateDraft
	model kvstore.KVStore
}

var _ rapid.StateMachine = &stateSM{}

// State Machine initialization.
func newStateSM() *stateSM {
	sm := new(stateSM)
	sm.store = statetest.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB())
	sm.draft = sm.store.NewOriginStateDraft()
	sm.model = mapdb.NewMapDB()
	return sm
}

// Action: Set a value for the KV store.
func (sm *stateSM) KVSet(t *rapid.T) {
	keyB := rapid.Byte().Draw(t, "key")
	valB := rapid.Byte().Draw(t, "val")
	sm.draft.Set(kv.Key([]byte{keyB}), []byte{valB})
	require.NoError(t, sm.model.Set([]byte{keyB}, []byte{valB}))
}

// Action: Set a value for the KV store (a longer slice).
func (sm *stateSM) KVSetSlices(t *rapid.T) {
	keyBin := rapid.SliceOfBytesMatching(".+").Draw(t, "key")
	valBin := rapid.SliceOfBytesMatching(".+").Draw(t, "val") // Nil values are not supported.
	sm.draft.Set(kv.Key(keyBin), valBin)
	require.NoError(t, sm.model.Set(keyBin, valBin))
}

// Action: Delete a value from the KV store.
func (sm *stateSM) KVDel(t *rapid.T) {
	keyB := rapid.Byte().Draw(t, "key")
	sm.draft.Del(kv.Key([]byte{keyB}))
	require.NoError(t, sm.model.Delete([]byte{keyB}))
}

// Action: Commit a block, start new empty draft.
func (sm *stateSM) CommitAddEmpty(t *rapid.T) {
	var err error
	block, _, _ := lo.Must3(sm.store.Commit(sm.draft))
	//
	// Validate, if the committed state is correct.
	blockState, err := sm.store.StateByTrieRoot(block.TrieRoot())
	require.NoError(t, err)
	sm.checkStateReaderMatchesModel(t, blockState)
	//
	// Proceed to the next transition.
	sm.draft, err = sm.store.NewEmptyStateDraft(block.L1Commitment())
	require.NoError(t, err)
}

// Invariants to check.
func (sm *stateSM) Check(t *rapid.T) {
	sm.checkStateReaderMatchesModel(t, sm.draft)
}

// Property: the state and the model DB have to have the same keys/values.
func (sm *stateSM) checkStateReaderMatchesModel(t *rapid.T, reader kv.KVStoreReader) {
	require.NoError(t, sm.model.Iterate(kvstore.EmptyPrefix, func(key, value kvstore.Value) bool {
		draftHasVal := reader.Has(kv.Key(key))
		require.True(t, draftHasVal, "Should have key %v", key)
		draftValue := reader.Get(kv.Key(key))
		require.Equal(t, value, draftValue, "Values for key %v should be equal", key)
		return true
	}))
	reader.Iterate(kv.EmptyPrefix, func(key kv.Key, value kvstore.Value) bool {
		modelHasVal, err := sm.model.Has([]byte(key))
		require.NoError(t, err)
		require.True(t, modelHasVal, "Should have key %v", key)
		modelValue, err := sm.model.Get([]byte(key))
		require.NoError(t, err)
		require.Equal(t, value, modelValue, "Values for key %v should be equal", key)
		return true
	})
}

func TestRapid(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		sm := newStateSM()
		t.Repeat(rapid.StateMachineActions(sm))
	})
}

func TestRapidReproduced(t *testing.T) {
	var err error
	store := statetest.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB())
	draft := store.NewOriginStateDraft()
	draft.Set(kv.Key([]byte{0}), []byte{0})
	draft.Set(kv.Key([]byte{1}), []byte{0})
	draft.Set(kv.Key([]byte{0x10}), []byte{0})
	//
	block, _, _ := lo.Must3(store.Commit(draft))
	blockState, err := store.StateByTrieRoot(block.TrieRoot())
	require.NoError(t, err)
	//
	check := func(b byte) {
		keyBin := []byte{b}
		key := kv.Key(keyBin)
		has := blockState.Has(key)
		require.True(t, has)
		val := blockState.Get(key)
		require.Equal(t, []byte{0}, val, "values equal for key %v", keyBin)
	}
	check(0)
	check(1)
	check(0x10)
}

func TestRapidReproduced2(t *testing.T) {
	store := statetest.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB())
	draft := store.NewOriginStateDraft()
	draft.Set(kv.Key([]byte{0x2}), []byte{0x1})
	draft.Set(kv.Key([]byte{0x7}), []byte{0x1})

	block, _, _ := lo.Must3(store.Commit(draft))
	root1 := block.TrieRoot()
	blockState, err := store.StateByTrieRoot(block.TrieRoot())
	t.Log(block.TrieRoot())
	require.NoError(t, err)

	require.Equal(t, blockState.Get(kv.Key([]byte{0x2})), []byte{0x1})
	require.Equal(t, blockState.Get(kv.Key([]byte{0x7})), []byte{0x1})

	//
	// Proceed to the next transition.
	draft, err = store.NewEmptyStateDraft(block.L1Commitment())
	require.NoError(t, err)

	draft.Set(kv.Key([]byte{0x2}), []byte{0x0})
	draft.Set(kv.Key([]byte{0x7}), []byte{0x1})

	block, _, _ = lo.Must3(store.Commit(draft))
	require.NotEqualValues(t, root1, block.TrieRoot())
}
