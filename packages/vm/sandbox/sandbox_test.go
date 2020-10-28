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
	s := stateWrapper{
		virtualState: state.NewVirtualState(db, &chainID),
		stateUpdate:  state.NewStateUpdate(nil),
	}

	s.Set("x", []byte{1})
	v, err := s.Get("x")

	assert.NoError(t, err)
	assert.Equal(t, []byte{1}, v)

	s.Del("x")
	v, err = s.Get("x")

	assert.NoError(t, err)
	assert.Nil(t, v)

	s.Set("x", []byte{2})
	s.Set("x", []byte{3})
	v, err = s.Get("x")

	assert.NoError(t, err)
	assert.Equal(t, []byte{3}, v)
}
