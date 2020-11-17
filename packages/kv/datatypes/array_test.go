package datatypes

import (
	"testing"

	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/stretchr/testify/assert"
)

func TestBasicArray(t *testing.T) {
	vars := dict.New()
	arr, err := NewArray(vars, "testArray")
	assert.NoError(t, err)

	d1 := []byte("datum1")
	d2 := []byte("datum2")
	d3 := []byte("datum3")
	d4 := []byte("datum4")

	arr.Push(d1)
	assert.EqualValues(t, 1, arr.Len())
	v, err := arr.GetAt(0)
	assert.NoError(t, err)
	assert.EqualValues(t, d1, v)
	v, err = arr.GetAt(1)
	assert.Error(t, err)
	assert.Nil(t, v)

	arr.Push(d2)
	assert.EqualValues(t, 2, arr.Len())

	arr.Push(d3)
	assert.EqualValues(t, 3, arr.Len())

	arr.Push(d4)
	assert.EqualValues(t, 4, arr.Len())

	arr2, err := NewArray(vars, "testArray2")
	assert.NoError(t, err)
	assert.EqualValues(t, 0, arr2.Len())

	arr2.Extend(arr)
	assert.EqualValues(t, arr.Len(), arr2.Len())

	arr2.Push(d4)
	assert.EqualValues(t, arr.Len()+1, arr2.Len())

	assert.Panics(t, func() {
		NewMustArray(arr2).GetAt(arr2.Len())
	})
}
