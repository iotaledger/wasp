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
	sm.store = state.NewStore(mapdb.NewMapDB())
	sm.draft = sm.store.NewOriginStateDraft()
	sm.model = mapdb.NewMapDB()
}

// Action: Set a value for the KV store.
func (sm *stateSM) KVSet(t *rapid.T) {
	keyB := rapid.Byte().Draw(t, "key")
	valB := rapid.Byte().Draw(t, "val")
	sm.draft.Set(kv.Key([]byte{keyB}), []byte{valB})
	sm.model.Set([]byte{keyB}, []byte{valB})
}

// Action: Commit a block.
func (sm *stateSM) Commit(t *rapid.T) {
	var err error
	block := sm.store.Commit(sm.draft)
	//
	// Validate, if the committed state is correct.
	blockState, err := sm.store.StateByTrieRoot(block.TrieRoot())
	require.NoError(t, err)
	sm.propStateReaderMatchesModel(t, blockState)
	//
	// Proceed to the next transition.
	sm.draft, err = sm.store.NewEmptyStateDraft(block.L1Commitment())
	require.NoError(t, err)
}

// Invariants to check.
func (sm *stateSM) Check(t *rapid.T) {
	sm.propStateReaderMatchesModel(t, sm.draft)
}

// Property: the state and the model DB have to have the same keys/values.
func (sm *stateSM) propStateReaderMatchesModel(t *rapid.T, reader kv.KVStoreReader) {
	require.NoError(t, sm.model.Iterate(kvstore.EmptyPrefix, func(key, value kvstore.Value) bool {
		draftValue, err := sm.draft.Get(kv.Key(key))
		require.NoError(t, err)
		require.Equal(t, draftValue, value)
		return true
	}))
	require.NoError(t, sm.draft.Iterate(kv.EmptyPrefix, func(key kv.Key, value kvstore.Value) bool {
		modelValue, err := sm.model.Get([]byte(key))
		require.NoError(t, err)
		require.Equal(t, modelValue, value)
		return true
	}))
}

func TestRapid(t *testing.T) {
	rapid.Check(t, rapid.Run[*stateSM]())
}
