package collections

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/kv/dict"
)

func TestBasicMap(t *testing.T) {
	vars := dict.New()
	m := NewMap(vars, "testMap")

	assert.Zero(t, m.MustLen())

	k1 := []byte("k1")
	k2 := []byte("k2")
	k3 := []byte("k3")
	k4 := []byte("") // empty key
	v1 := []byte("datum1")
	v2 := []byte("datum2")
	v3 := []byte("datum3")
	v4 := []byte("datum4")

	m.MustSetAt(k1, v1)
	ok := m.MustHasAt(k1)
	assert.True(t, ok)

	ok = m.MustHasAt(k2)
	assert.False(t, ok)

	ok = m.MustHasAt(k3)
	assert.False(t, ok)
	assert.EqualValues(t, 1, m.MustLen())

	ok = m.MustHasAt(k4)
	assert.False(t, ok)
	assert.EqualValues(t, 1, m.MustLen())

	v := m.MustGetAt(k1)
	assert.EqualValues(t, v1, v)

	v = m.MustGetAt(k2)
	assert.Nil(t, v)

	m.MustSetAt(k2, v2)
	m.MustSetAt(k3, v3)

	ok = m.MustHasAt(k1)
	assert.True(t, ok)

	ok = m.MustHasAt(k2)
	assert.True(t, ok)

	ok = m.MustHasAt(k3)
	assert.True(t, ok)

	assert.EqualValues(t, 3, m.MustLen())

	v = m.MustGetAt(k2)
	assert.EqualValues(t, v2, v)

	v = m.MustGetAt(k3)
	assert.EqualValues(t, v3, v)

	m.MustDelAt(k2)

	ok = m.MustHasAt(k1)
	assert.True(t, ok)

	ok = m.MustHasAt(k2)
	assert.False(t, ok)

	ok = m.MustHasAt(k3)
	assert.True(t, ok)

	assert.EqualValues(t, 2, m.MustLen())

	v = m.MustGetAt(k2)
	assert.Nil(t, v)

	v = m.MustGetAt(k3)
	assert.EqualValues(t, v3, v)

	m.MustSetAt(k4, v4)
	v = m.MustGetAt(k4)
	assert.EqualValues(t, v, v4)

	m.MustDelAt(k4)
	v, err := m.GetAt(k4)
	assert.NoError(t, err)
	assert.Nil(t, v)
}

func TestIterate(t *testing.T) {
	vars := dict.New()
	m := NewMap(vars, "testMap")

	assert.Zero(t, m.MustLen())

	kv := map[string]string{
		"k1": "v1",
		"k2": "v2",
		"k3": "v3",
		"":   "empty",
	}
	for k, v := range kv {
		m.MustSetAt([]byte(k), []byte(v))
	}
	m.MustIterate(func(k []byte, v []byte) bool {
		assert.EqualValues(t, kv[string(k)], v)
		return true
	})
	m.MustDelAt([]byte("k1"))
	m.MustIterate(func(k []byte, v []byte) bool {
		assert.NotEqualValues(t, k, "k1")
		assert.EqualValues(t, kv[string(k)], v)
		return true
	})
	m.MustDelAt([]byte(""))
	m.MustIterate(func(k []byte, v []byte) bool {
		assert.NotEqualValues(t, k, "k1")
		assert.NotEqualValues(t, k, "")
		assert.EqualValues(t, kv[string(k)], v)
		return true
	})
}

func TestMapConcurrentAccess(t *testing.T) {
	vars := dict.New()
	m1 := NewMap(vars, "testMap")
	m2 := NewMap(vars, "testMap")

	m1.MustSetAt([]byte{1}, []byte{1})
	require.EqualValues(t, m1.MustLen(), 1)
	require.EqualValues(t, m2.MustLen(), 1)

	m2.MustDelAt([]byte{1})
	require.EqualValues(t, m1.MustLen(), 0)
	require.EqualValues(t, m2.MustLen(), 0)
}
