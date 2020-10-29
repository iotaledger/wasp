package sandbox

import (
	"testing"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/stretchr/testify/assert"
)

func TestSetThenGet(t *testing.T) {
	db := mapdb.NewMapDB()

	chainID := coretypes.ChainID{1, 3, 3, 7}

	virtualState := state.NewVirtualState(db, &chainID)
	stateUpdate := state.NewStateUpdate(nil)

	s := stateWrapper{
		contractIndex: 2,
		virtualState:  virtualState.Variables().Codec(),
		stateUpdate:   stateUpdate,
	}

	// contract sets variable x
	s.Set("x", []byte{1})

	// contract gets variable x
	v, err := s.Get("x")
	assert.NoError(t, err)
	assert.Equal(t, []byte{1}, v)

	// mutation is in stateUpdate, prefixed by the contract id
	assert.Equal(t, []byte{1}, stateUpdate.Mutations().Latest("\x02\x00x").Value())

	// mutation is not committed to the virtual state
	v, err = virtualState.Variables().Get("\x02\x00x")
	assert.NoError(t, err)
	assert.Nil(t, v)

	// contract deletes variable x
	s.Del("x")

	// contract sees variable x does not exist
	v, err = s.Get("x")
	assert.NoError(t, err)
	assert.Nil(t, v)

	// contract makes several writes to same variable, gets the latest value
	s.Set("x", []byte{2})
	s.Set("x", []byte{3})
	v, err = s.Get("x")

	assert.NoError(t, err)
	assert.Equal(t, []byte{3}, v)

	// all changes are mutations in stateUpdate
	assert.Equal(t, 4, stateUpdate.Mutations().Len())
}
