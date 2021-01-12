package collections

import (
	"testing"

	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/stretchr/testify/assert"
)

func TestBasicArray(t *testing.T) {
	vars := dict.New()
	arr := NewArray(vars, "testArray")

	d1 := []byte("datum1")
	d2 := []byte("datum2")
	d3 := []byte("datum3")
	d4 := []byte("datum4")

	arr.MustPush(d1)
	assert.EqualValues(t, 1, arr.MustLen())
	v := arr.MustGetAt(0)
	assert.EqualValues(t, d1, v)
	assert.Panics(t, func() {
		arr.MustGetAt(1)
	})

	arr.MustPush(d2)
	assert.EqualValues(t, 2, arr.MustLen())

	arr.MustPush(d3)
	assert.EqualValues(t, 3, arr.MustLen())

	arr.MustPush(d4)
	assert.EqualValues(t, 4, arr.MustLen())

	arr2 := NewArray(vars, "testArray2")
	assert.EqualValues(t, 0, arr2.MustLen())

	arr2.MustExtend(arr.Immutable())
	assert.EqualValues(t, arr.MustLen(), arr2.MustLen())

	arr2.MustPush(d4)
	assert.EqualValues(t, arr.MustLen()+1, arr2.MustLen())
}

func TestConcurrentAccess(t *testing.T) {
	vars := dict.New()
	a1 := NewArray(vars, "test")
	a2 := NewArray(vars, "test")

	a1.MustPush([]byte{1})
	assert.EqualValues(t, a1.MustLen(), 1)
	assert.EqualValues(t, a2.MustLen(), 1)
}
