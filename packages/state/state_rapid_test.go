package state_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	"github.com/iotaledger/hive.go/core/kvstore"
	"github.com/iotaledger/hive.go/core/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
)

type stateSM struct {
	store state.Store
	draft state.StateDraft
	model kvstore.KVStore
}

var _ rapid.StateMachine = &stateSM{}

// State Machine initialization.
func (sm *stateSM) Init(t *rapid.T) {
	// fmt.Printf("XXX: Init\n")
	sm.store = state.NewStore(mapdb.NewMapDB())
	sm.draft = sm.store.NewOriginStateDraft()
	sm.model = mapdb.NewMapDB()
}

// Action: Set a value for the KV store.
func (sm *stateSM) KVSet(t *rapid.T) {
	keyB := rapid.Byte().Draw(t, "key")
	valB := rapid.Byte().Draw(t, "val")
	// fmt.Printf("XXX: KVSet %v=>%v\n", keyB, valB)
	sm.draft.Set(kv.Key([]byte{keyB}), []byte{valB})
	require.NoError(t, sm.model.Set([]byte{keyB}, []byte{valB}))
}

// Action: Commit a block, start new empty draft.
func (sm *stateSM) CommitAddEmpty(t *rapid.T) {
	// fmt.Printf("XXX: CommitAddEmpty\n")
	var err error
	block := sm.store.Commit(sm.draft)
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

// TODO: Action: Commit a block, start new draft with the common fields.
// func (sm *stateSM) CommitAddDraft(t *rapid.T) {
// 	var err error
// 	block := sm.store.Commit(sm.draft)
// 	//
// 	// Validate, if the committed state is correct.
// 	blockState, err := sm.store.StateByTrieRoot(block.TrieRoot())
// 	require.NoError(t, err)
// 	sm.propStateReaderMatchesModel(t, blockState)
// 	//
// 	// Proceed to the next transition.
// 	timestamp := rapid.Int64().Draw(t, "timestamp")
// 	sm.draft, err = sm.store.NewStateDraft(time.UnixMilli(timestamp), block.L1Commitment())
// 	require.NoError(t, err)
// }

// Invariants to check.
func (sm *stateSM) Check(t *rapid.T) {
	sm.checkStateReaderMatchesModel(t, sm.draft)
}

// Property: the state and the model DB have to have the same keys/values.
func (sm *stateSM) checkStateReaderMatchesModel(t *rapid.T, reader kv.KVStoreReader) {
	require.NoError(t, sm.model.Iterate(kvstore.EmptyPrefix, func(key, value kvstore.Value) bool {
		draftHasVal, err := reader.Has(kv.Key(key))
		require.NoError(t, err)
		require.True(t, draftHasVal, "Should have key %v", key)
		draftValue, err := reader.Get(kv.Key(key))
		require.NoError(t, err)
		require.Equal(t, value, draftValue, "Values for key %v should be equal", key)
		return true
	}))
	require.NoError(t, reader.Iterate(kv.EmptyPrefix, func(key kv.Key, value kvstore.Value) bool {
		modelHasVal, err := sm.model.Has([]byte(key))
		require.NoError(t, err)
		require.True(t, modelHasVal, "Should have key %v", key)
		modelValue, err := sm.model.Get([]byte(key))
		require.NoError(t, err)
		require.Equal(t, value, modelValue, "Values for key %v should be equal", key)
		return true
	}))
}

func TestRapid(t *testing.T) {
	rapid.Check(t, rapid.Run[*stateSM]())
}

// === RUN   TestRapid
//
//	state_rapid_test.go:97: [rapid] failed after 3 tests: (*T).FailNow() called
//	    To reproduce, specify -run="TestRapid" -rapid.failfile="testdata/rapid/TestRapid/TestRapid-20221207135045-108636.fail" (or -rapid.seed=16363909978550960135)
//	    Failed test output:
//	state_rapid_test.go:97: [rapid] draw action: "KVSet"
//	state_rapid_test.go:32: [rapid] draw key: 0x0
//	state_rapid_test.go:33: [rapid] draw val: 0x0
//	state_rapid_test.go:97: [rapid] draw action: "CommitAddEmpty"
//	state_rapid_test.go:97: [rapid] draw action: "KVSet"
//	state_rapid_test.go:32: [rapid] draw key: 0x1
//	state_rapid_test.go:33: [rapid] draw val: 0x0
//	state_rapid_test.go:97: [rapid] draw action: "CommitAddEmpty"
//	state_rapid_test.go:79:
//	            Error Trace:    wasp/packages/state/state_rapid_test.go:79
//	                            wasp/packages/state/synced_map.go:79
//	                            wasp/packages/state/mapdb.go:59
//	                            wasp/packages/state/state_rapid_test.go:76
//	                            wasp/packages/state/state_rapid_test.go:46
//	                            wasp/packages/state/statemachine.go:174
//	                            wasp/packages/state/statemachine.go:149
//	                            wasp/packages/state/statemachine.go:83
//	                            wasp/packages/state/engine.go:319
//	                            wasp/packages/state/engine.go:201
//	                            wasp/packages/state/engine.go:93
//	                            wasp/packages/state/state_rapid_test.go:97
//	            Error:          Should be true
//	            Test:           TestRapid
//	            Messages:       Should have key [0]
//
// --- FAIL: TestRapid (4.17s)
// FAIL
// FAIL    github.com/iotaledger/wasp/packages/state       4.189s
// FAIL
func TestRapidReproduced(t *testing.T) { // TODO: Fails to reproduce....
	cmp := func(model kvstore.KVStore, reader kv.KVStoreReader) {
		require.NoError(t, model.Iterate(kvstore.EmptyPrefix, func(key, value kvstore.Value) bool {
			draftHasVal, err := reader.Has(kv.Key(key))
			require.NoError(t, err)
			require.True(t, draftHasVal, "Should have key %v", key)
			draftValue, err := reader.Get(kv.Key(key))
			require.NoError(t, err)
			require.Equal(t, value, draftValue, "Values for key %v should be equal", key)
			return true
		}))
		require.NoError(t, reader.Iterate(kv.EmptyPrefix, func(key kv.Key, value kvstore.Value) bool {
			modelHasVal, err := model.Has([]byte(key))
			require.NoError(t, err)
			require.True(t, modelHasVal, "Should have key %v", key)
			modelValue, err := model.Get([]byte(key))
			require.NoError(t, err)
			require.Equal(t, value, modelValue, "Values for key %v should be equal", key)
			return true
		}))
	}

	var err error
	// XXX: Init
	store := state.NewStore(mapdb.NewMapDB())
	draft := store.NewOriginStateDraft()
	model := mapdb.NewMapDB()
	cmp(model, draft)
	// XXX: KVSet 0=>0
	draft.Set(kv.Key([]byte{0}), []byte{0})
	model.Set([]byte{0}, []byte{0})
	cmp(model, draft)
	// XXX: CommitAddEmpty
	block := store.Commit(draft)
	blockState, err := store.StateByTrieRoot(block.TrieRoot())
	require.NoError(t, err)
	cmp(model, blockState)
	draft, err = store.NewEmptyStateDraft(block.L1Commitment())
	require.NoError(t, err)
	cmp(model, draft)
	// XXX: KVSet 1=>0
	draft.Set(kv.Key([]byte{1}), []byte{0})
	model.Set([]byte{1}, []byte{0})
	cmp(model, draft)
	// XXX: CommitAddEmpty
	block = store.Commit(draft)
	blockState, err = store.StateByTrieRoot(block.TrieRoot())
	require.NoError(t, err)
	cmp(model, blockState)
	draft, err = store.NewEmptyStateDraft(block.L1Commitment())
	require.NoError(t, err)
	cmp(model, draft)
	//////
	require.NoError(t, err)
	has, err := blockState.Has(kv.Key([]byte{0}))
	require.NoError(t, err)
	require.True(t, has)
	has, err = blockState.Has(kv.Key([]byte{1}))
	require.NoError(t, err)
	require.True(t, has)
}

func TestRapidReproduced2(t *testing.T) {
	var err error
	// XXX: Init
	store := state.NewStore(mapdb.NewMapDB())
	draft := store.NewOriginStateDraft()
	model := mapdb.NewMapDB()
	// XXX: KVSet 0x0=>0
	draft.Set(kv.Key([]byte{0}), []byte{0})
	model.Set([]byte{0x0}, []byte{0})
	// XXX: KVSet 0x1=>0
	draft.Set(kv.Key([]byte{0}), []byte{0})
	model.Set([]byte{0x1}, []byte{0})
	// XXX: KVSet 0x16=>0
	draft.Set(kv.Key([]byte{0}), []byte{0})
	model.Set([]byte{0x16}, []byte{0})
	// XXX: CommitAddEmpty
	block := store.Commit(draft)
	blockState, err := store.StateByTrieRoot(block.TrieRoot())
	require.NoError(t, err)
	//////
	require.NoError(t, err)
	has, err := blockState.Has(kv.Key([]byte{0}))
	require.NoError(t, err)
	require.True(t, has)
	has, err = blockState.Has(kv.Key([]byte{1}))
	require.NoError(t, err)
	require.True(t, has)
	has, err = blockState.Has(kv.Key([]byte{16}))
	require.NoError(t, err)
	require.True(t, has)
}
