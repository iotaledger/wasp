package collections

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/kv/dict"
)

func TestBasicArray(t *testing.T) {
	vars := dict.New()
	arr := NewArray(vars, "testArray")

	d1 := []byte("datum1")
	d2 := []byte("datum2")
	d3 := []byte("datum3")
	d4 := []byte("datum4")

	arr.Push(d1)
	require.EqualValues(t, 1, arr.Len())
	v := arr.GetAt(0)
	require.EqualValues(t, d1, v)
	require.Panics(t, func() {
		arr.GetAt(1)
	})

	arr.Push(d2)
	require.EqualValues(t, 2, arr.Len())

	arr.Push(d3)
	require.EqualValues(t, 3, arr.Len())

	arr.Push(d4)
	require.EqualValues(t, 4, arr.Len())

	arr2 := NewArray(vars, "testArray2")
	require.EqualValues(t, 0, arr2.Len())

	arr2.Extend(arr.Immutable())
	require.EqualValues(t, arr.Len(), arr2.Len())

	arr2.Push(d4)
	require.EqualValues(t, arr.Len()+1, arr2.Len())
}

func TestConcurrentAccessArray(t *testing.T) {
	vars := dict.New()
	a1 := NewArray(vars, "test")
	a2 := NewArray(vars, "test")

	a1.Push([]byte{1})
	require.EqualValues(t, a1.Len(), 1)
	require.EqualValues(t, a2.Len(), 1)
}
