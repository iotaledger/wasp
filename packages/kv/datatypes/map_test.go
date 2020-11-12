package datatypes

import (
	"testing"

	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/stretchr/testify/assert"
)

func TestBasicMap(t *testing.T) {
	vars := dict.New()
	m, err := NewMap(vars, "testMap")
	assert.NoError(t, err)

	assert.Zero(t, m.Len())

	k1 := []byte("k1")
	k2 := []byte("k2")
	k3 := []byte("k3")
	v1 := []byte("datum1")
	v2 := []byte("datum2")
	v3 := []byte("datum3")

	m.SetAt(k1, v1)
	ok, err := m.HasAt(k1)
	assert.True(t, ok)
	assert.NoError(t, err)

	ok, err = m.HasAt(k2)
	assert.False(t, ok)
	assert.NoError(t, err)

	ok, err = m.HasAt(k3)
	assert.False(t, ok)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, m.Len())

	v, err := m.GetAt(k1)
	assert.EqualValues(t, v1, v)

	v, err = m.GetAt(k2)
	assert.NoError(t, err)
	assert.Nil(t, v)

	m.SetAt(k2, v2)
	m.SetAt(k3, v3)

	ok, err = m.HasAt(k1)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = m.HasAt(k2)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = m.HasAt(k3)
	assert.NoError(t, err)
	assert.True(t, ok)

	assert.EqualValues(t, 3, m.Len())

	v, err = m.GetAt(k2)
	assert.NoError(t, err)
	assert.EqualValues(t, v2, v)

	v, err = m.GetAt(k3)
	assert.NoError(t, err)
	assert.EqualValues(t, v3, v)

	err = m.DelAt(k2)
	assert.NoError(t, err)

	ok, err = m.HasAt(k1)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = m.HasAt(k2)
	assert.NoError(t, err)
	assert.False(t, ok)

	ok, err = m.HasAt(k3)
	assert.NoError(t, err)
	assert.True(t, ok)

	assert.EqualValues(t, 2, m.Len())

	v, err = m.GetAt(k2)
	assert.NoError(t, err)
	assert.Nil(t, v)

	v, err = m.GetAt(k3)
	assert.NoError(t, err)
	assert.EqualValues(t, v3, v)
}
