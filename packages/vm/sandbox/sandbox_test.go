package sandbox

import (
	"testing"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/stretchr/testify/assert"
)

func TestSetThenGet(t *testing.T) {
	addr := address.Random()
	s := stateWrapper{
		virtualState: state.NewEmptyVirtualState(&addr),
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
