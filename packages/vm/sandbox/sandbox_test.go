package sandbox

import (
	"testing"

	"github.com/iotaledger/wasp/packages/state"
	"github.com/stretchr/testify/assert"
)

func TestSetThenGet(t *testing.T) {
	s := stateWrapper{
		virtualState: state.NewVirtualState(nil),
		stateUpdate:  state.NewStateUpdate(nil),
	}

	s.Set("x", []byte{1})
	v, ok := s.Get("x")

	assert.True(t, ok)
	assert.Equal(t, []byte{1}, v)

	s.Del("x")
	v, ok = s.Get("x")

	assert.False(t, ok)

	s.Set("x", []byte{2})
	s.Set("x", []byte{3})
	v, ok = s.Get("x")

	assert.True(t, ok)
	assert.Equal(t, []byte{3}, v)
}
