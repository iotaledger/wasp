package datatypes

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/stretchr/testify/assert"
)

func TestBasicMap(t *testing.T) {
	vars := dict.New()
	m := NewMustMap(vars, "testMap")

	assert.Zero(t, m.Len())

	k1 := []byte("k1")
	k2 := []byte("k2")
	k3 := []byte("k3")
	v1 := []byte("datum1")
	v2 := []byte("datum2")
	v3 := []byte("datum3")

	m.SetAt(k1, v1)
	ok := m.HasAt(k1)
	assert.True(t, ok)

	ok = m.HasAt(k2)
	assert.False(t, ok)

	ok = m.HasAt(k3)
	assert.False(t, ok)
	assert.EqualValues(t, 1, m.Len())

	v := m.GetAt(k1)
	assert.EqualValues(t, v1, v)

	v = m.GetAt(k2)
	assert.Nil(t, v)

	m.SetAt(k2, v2)
	m.SetAt(k3, v3)

	ok = m.HasAt(k1)
	assert.True(t, ok)

	ok = m.HasAt(k2)
	assert.True(t, ok)

	ok = m.HasAt(k3)
	assert.True(t, ok)

	assert.EqualValues(t, 3, m.Len())

	v = m.GetAt(k2)
	assert.EqualValues(t, v2, v)

	v = m.GetAt(k3)
	assert.EqualValues(t, v3, v)

	m.DelAt(k2)

	ok = m.HasAt(k1)
	assert.True(t, ok)

	ok = m.HasAt(k2)
	assert.False(t, ok)

	ok = m.HasAt(k3)
	assert.True(t, ok)

	assert.EqualValues(t, 2, m.Len())

	v = m.GetAt(k2)
	assert.Nil(t, v)

	v = m.GetAt(k3)
	assert.EqualValues(t, v3, v)
}

func TestIterate(t *testing.T) {
	vars := dict.New()
	m := NewMustMap(vars, "testMap")

	assert.Zero(t, m.Len())

	keys := []string{"k1", "k2", "k3"}
	values := []string{"v1", "v2", "v3"}
	for i, key := range keys {
		m.SetAt([]byte(key), []byte(values[i]))
	}
	m.Iterate(func(elemKey []byte, value []byte) bool {
		t.Logf("key '%s' value '%s'", string(elemKey), string(value))
		return true
	})
	t.Logf("---------------------")
	m.SetAt([]byte("k4"), []byte("v4"))
	m.Iterate(func(elemKey []byte, value []byte) bool {
		t.Logf("key '%s' value '%s'", string(elemKey), string(value))
		return true
	})
}

func TestConcurrentAccess(t *testing.T) {
	vars := dict.New()
	m1 := NewMustMap(vars, "testMap")
	m2 := NewMustMap(vars, "testMap")

	m1.SetAt([]byte{1}, []byte{1})
	require.EqualValues(t, m1.Len(), 1)
	require.EqualValues(t, m2.Len(), 1)

	m2.DelAt([]byte{1})
	require.EqualValues(t, m1.Len(), 0)
	require.EqualValues(t, m2.Len(), 0)
}
