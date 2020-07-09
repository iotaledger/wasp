package kv

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBasicDic(t *testing.T) {
	vars := NewMap()
	dict := newDict(vars, "testDict")

	assert.Zero(t, dict.Len())

	k1 := []byte("k1")
	k2 := []byte("k2")
	k3 := []byte("k3")
	v1 := []byte("datum1")
	v2 := []byte("datum2")
	v3 := []byte("datum3")

	dict.SetAt(k1, v1)
	assert.True(t, dict.HasAt(k1))
	assert.False(t, dict.HasAt(k2))
	assert.False(t, dict.HasAt(k3))
	assert.EqualValues(t, 1, dict.Len())

	v, ok := dict.GetAt(k1)
	assert.True(t, ok)
	assert.EqualValues(t, v1, v)

	v, ok = dict.GetAt(k2)
	assert.False(t, ok)
	assert.Nil(t, v)

	dict.SetAt(k2, v2)
	dict.SetAt(k3, v3)
	assert.True(t, dict.HasAt(k1))
	assert.True(t, dict.HasAt(k2))
	assert.True(t, dict.HasAt(k3))
	assert.EqualValues(t, 3, dict.Len())

	dict.DelAt(k2)
	assert.True(t, dict.HasAt(k1))
	assert.False(t, dict.HasAt(k2))
	assert.True(t, dict.HasAt(k3))
	assert.EqualValues(t, 2, dict.Len())

}
