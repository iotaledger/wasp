package kv

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBasicDict(t *testing.T) {
	vars := NewMap()
	dict, err := newDictionary(vars, "testDict")
	assert.NoError(t, err)

	assert.Zero(t, dict.Len())

	k1 := []byte("k1")
	k2 := []byte("k2")
	k3 := []byte("k3")
	v1 := []byte("datum1")
	v2 := []byte("datum2")
	v3 := []byte("datum3")

	dict.SetAt(k1, v1)
	ok, err := dict.HasAt(k1)
	assert.True(t, ok)
	assert.NoError(t, err)

	ok, err = dict.HasAt(k2)
	assert.False(t, ok)
	assert.NoError(t, err)

	ok, err = dict.HasAt(k3)
	assert.False(t, ok)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, dict.Len())

	v, err := dict.GetAt(k1)
	assert.EqualValues(t, v1, v)

	v, err = dict.GetAt(k2)
	assert.NoError(t, err)
	assert.Nil(t, v)

	dict.SetAt(k2, v2)
	dict.SetAt(k3, v3)

	ok, err = dict.HasAt(k1)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = dict.HasAt(k2)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = dict.HasAt(k3)
	assert.NoError(t, err)
	assert.True(t, ok)

	assert.EqualValues(t, 3, dict.Len())

	v, err = dict.GetAt(k2)
	assert.NoError(t, err)
	assert.EqualValues(t, v2, v)

	v, err = dict.GetAt(k3)
	assert.NoError(t, err)
	assert.EqualValues(t, v3, v)

	err = dict.DelAt(k2)
	assert.NoError(t, err)

	ok, err = dict.HasAt(k1)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = dict.HasAt(k2)
	assert.NoError(t, err)
	assert.False(t, ok)

	ok, err = dict.HasAt(k3)
	assert.NoError(t, err)
	assert.True(t, ok)

	assert.EqualValues(t, 2, dict.Len())

	v, err = dict.GetAt(k2)
	assert.NoError(t, err)
	assert.Nil(t, v)

	v, err = dict.GetAt(k3)
	assert.NoError(t, err)
	assert.EqualValues(t, v3, v)
}
