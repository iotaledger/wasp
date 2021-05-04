package vmcontext

import (
	"github.com/iotaledger/wasp/packages/dbprovider"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"testing"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/stretchr/testify/assert"
)

func TestSetThenGet(t *testing.T) {
	dbp := dbprovider.NewInMemoryDBProvider(testlogger.NewLogger(t))
	chainID := coretypes.RandomChainID([]byte("hmm"))
	virtualState, err := state.CreateOriginState(dbp, chainID)

	stateUpdate := state.NewStateUpdate()
	hname := coretypes.Hn("test")

	s := newStateWrapper(hname, virtualState, stateUpdate)

	subpartitionedKey := kv.Key(hname.Bytes()) + "x"

	// contract sets variable x
	s.Set("x", []byte{1})
	assert.Equal(t, map[kv.Key][]byte{subpartitionedKey: []byte{1}}, stateUpdate.Mutations().Sets)
	assert.Equal(t, map[kv.Key]struct{}{}, stateUpdate.Mutations().Dels)

	// contract gets variable x
	v, err := s.Get("x")
	assert.NoError(t, err)
	assert.Equal(t, []byte{1}, v)

	// mutation is in currentStateUpdate, prefixed by the contract id
	assert.Equal(t, []byte{1}, stateUpdate.Mutations().Sets[subpartitionedKey])

	// mutation is not committed to the virtual state
	v, err = virtualState.KVStore().Get(subpartitionedKey)
	assert.NoError(t, err)
	assert.Nil(t, v)

	// contract deletes variable x
	s.Del("x")
	assert.Equal(t, map[kv.Key][]byte{}, stateUpdate.Mutations().Sets)
	assert.Equal(t, map[kv.Key]struct{}{subpartitionedKey: struct{}{}}, stateUpdate.Mutations().Dels)

	// contract sees variable x does not exist
	v, err = s.Get("x")
	assert.NoError(t, err)
	assert.Nil(t, v)

	// contract makes several writes to same variable, gets the latest value
	s.Set("x", []byte{2})
	assert.Equal(t, map[kv.Key][]byte{subpartitionedKey: []byte{2}}, stateUpdate.Mutations().Sets)
	assert.Equal(t, map[kv.Key]struct{}{}, stateUpdate.Mutations().Dels)

	s.Set("x", []byte{3})
	assert.Equal(t, map[kv.Key][]byte{subpartitionedKey: []byte{3}}, stateUpdate.Mutations().Sets)
	assert.Equal(t, map[kv.Key]struct{}{}, stateUpdate.Mutations().Dels)

	v, err = s.Get("x")

	assert.NoError(t, err)
	assert.Equal(t, []byte{3}, v)
}

func TestIterate(t *testing.T) {
	dbp := dbprovider.NewInMemoryDBProvider(testlogger.NewLogger(t))
	chainID := coretypes.RandomChainID([]byte("hmm"))
	virtualState, err := state.CreateOriginState(dbp, chainID)

	stateUpdate := state.NewStateUpdate()
	hname := coretypes.Hn("test")

	s := newStateWrapper(hname, virtualState, stateUpdate)

	s.Set("xyz", []byte{1})

	err = s.Iterate("x", func(k kv.Key, v []byte) bool {
		assert.EqualValues(t, "xyz", string(k))
		return true
	})
	assert.NoError(t, err)
}
