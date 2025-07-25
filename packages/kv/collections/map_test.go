package collections

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/kv/dict"
)

func TestBasicMap(t *testing.T) {
	vars := dict.New()
	m := NewMap(vars, "testMap")

	require.Zero(t, m.Len())

	k1 := []byte("k1")
	k2 := []byte("k2")
	k3 := []byte("k3")
	k4 := []byte("") // empty key
	v1 := []byte("datum1")
	v2 := []byte("datum2")
	v3 := []byte("datum3")
	v4 := []byte("datum4")

	m.SetAt(k1, v1)
	ok := m.HasAt(k1)
	require.True(t, ok)

	ok = m.HasAt(k2)
	require.False(t, ok)

	ok = m.HasAt(k3)
	require.False(t, ok)
	require.EqualValues(t, 1, m.Len())

	ok = m.HasAt(k4)
	require.False(t, ok)
	require.EqualValues(t, 1, m.Len())

	v := m.GetAt(k1)
	require.EqualValues(t, v1, v)

	v = m.GetAt(k2)
	require.Nil(t, v)

	m.SetAt(k2, v2)
	m.SetAt(k3, v3)

	ok = m.HasAt(k1)
	require.True(t, ok)

	ok = m.HasAt(k2)
	require.True(t, ok)

	ok = m.HasAt(k3)
	require.True(t, ok)

	require.EqualValues(t, 3, m.Len())

	v = m.GetAt(k2)
	require.EqualValues(t, v2, v)

	v = m.GetAt(k3)
	require.EqualValues(t, v3, v)

	m.DelAt(k2)

	ok = m.HasAt(k1)
	require.True(t, ok)

	ok = m.HasAt(k2)
	require.False(t, ok)

	ok = m.HasAt(k3)
	require.True(t, ok)

	require.EqualValues(t, 2, m.Len())

	v = m.GetAt(k2)
	require.Nil(t, v)

	v = m.GetAt(k3)
	require.EqualValues(t, v3, v)

	m.SetAt(k4, v4)
	v = m.GetAt(k4)
	require.EqualValues(t, v, v4)

	m.DelAt(k4)
	v = m.GetAt(k4)
	require.Nil(t, v)
}

func TestIterate(t *testing.T) {
	vars := dict.New()
	m := NewMap(vars, "testMap")

	require.Zero(t, m.Len())

	kv := map[string]string{
		"k1": "v1",
		"k2": "v2",
		"k3": "v3",
		"":   "empty",
	}
	for k, v := range kv {
		m.SetAt([]byte(k), []byte(v))
	}
	m.Iterate(func(k []byte, v []byte) bool {
		require.EqualValues(t, kv[string(k)], v)
		return true
	})
	m.DelAt([]byte("k1"))
	m.Iterate(func(k []byte, v []byte) bool {
		require.NotEqualValues(t, k, "k1")
		require.EqualValues(t, kv[string(k)], v)
		return true
	})
	m.DelAt([]byte(""))
	m.Iterate(func(k []byte, v []byte) bool {
		require.NotEqualValues(t, k, "k1")
		require.NotEqualValues(t, k, "")
		require.EqualValues(t, kv[string(k)], v)
		return true
	})
}

func TestMapConcurrentAccess(t *testing.T) {
	vars := dict.New()
	m1 := NewMap(vars, "testMap")
	m2 := NewMap(vars, "testMap")

	m1.SetAt([]byte{1}, []byte{1})
	require.EqualValues(t, m1.Len(), 1)
	require.EqualValues(t, m2.Len(), 1)

	m2.DelAt([]byte{1})
	require.EqualValues(t, m1.Len(), 0)
	require.EqualValues(t, m2.Len(), 0)
}
