package collections

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/stretchr/testify/assert"
)

func TestBasicMap(t *testing.T) {
	vars := dict.New()
	m := NewMap(vars, "testMap")

	assert.Zero(t, m.MustLen())

	k1 := []byte("k1")
	k2 := []byte("k2")
	k3 := []byte("k3")
	v1 := []byte("datum1")
	v2 := []byte("datum2")
	v3 := []byte("datum3")

	m.MustSetAt(k1, v1)
	ok := m.MustHasAt(k1)
	assert.True(t, ok)

	ok = m.MustHasAt(k2)
	assert.False(t, ok)

	ok = m.MustHasAt(k3)
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
}

func TestIterate(t *testing.T) {
	vars := dict.New()
	m := NewMap(vars, "testMap")

	assert.Zero(t, m.MustLen())

	keys := []string{"k1", "k2", "k3"}
	values := []string{"v1", "v2", "v3"}
	for i, key := range keys {
		m.MustSetAt([]byte(key), []byte(values[i]))
	}
	m.MustIterate(func(elemKey []byte, value []byte) bool {
		t.Logf("key '%s' value '%s'", string(elemKey), string(value))
		return true
	})
	t.Logf("---------------------")
	m.MustSetAt([]byte("k4"), []byte("v4"))
	m.MustIterate(func(elemKey []byte, value []byte) bool {
		t.Logf("key '%s' value '%s'", string(elemKey), string(value))
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
