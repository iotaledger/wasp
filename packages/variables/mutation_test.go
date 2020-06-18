package variables

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyMutationSet(t *testing.T) {
	vars := New(nil)

	mset := NewMutationSet("k1", []byte("v1"))
	mset.Apply(vars)

	v, ok := vars.Get("k1")
	assert.Equal(t, true, ok)
	assert.Equal(t, []byte("v1"), v)
}

func TestApplyMutationDel(t *testing.T) {
	vars := New(nil)
	vars.Set("k1", []byte("v1"))

	mset := NewMutationDel("k1")
	mset.Apply(vars)

	_, ok := vars.Get("k1")
	assert.Equal(t, false, ok)
}
