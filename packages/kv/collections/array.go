package collections

import (
	"errors"
	"fmt"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

var ErrArrayOverflow = errors.New("Array overflow")

// For easy distinction between arrays and map collections
// we use '#' as separator for arrays and '.' for maps.
// Do not change this value unless you want to break how
// WasmLib maps these collections in the exact same way
const arrayElemKeyCode = byte('#')

func ArrayElemKey(name string, index uint32) kv.Key {
	key := append([]byte(name), arrayElemKeyCode)
	return kv.Key(append(key, rwutil.Size32ToBytes(index)...))
}

/////////////////////////////////  ArrayReadOnly  \\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\

// ArrayReadOnly provides read-only access to an Array in a kv.KVStoreReader.
type ArrayReadOnly struct {
	kvr  kv.KVStoreReader
	name string
}

func NewArrayReadOnly(kvReader kv.KVStoreReader, name string) *ArrayReadOnly {
	return &ArrayReadOnly{
		kvr:  kvReader,
		name: name,
	}
}

func (a *ArrayReadOnly) getArrayElemKey(index uint32) kv.Key {
	return ArrayElemKey(a.name, index)
}

func (a *ArrayReadOnly) GetAt(index uint32) []byte {
	length := a.Len()
	if index >= length {
		panic(fmt.Errorf("index %d out of range for array of len %d", index, length))
	}
	return a.kvr.Get(a.getArrayElemKey(index))
}

func (a *ArrayReadOnly) getSizeKey() kv.Key {
	return kv.Key(a.name)
}

// Len == 0/empty/non-existent are equivalent
func (a *ArrayReadOnly) Len() uint32 {
	v := a.kvr.Get(a.getSizeKey())
	if v == nil {
		return 0
	}
	return rwutil.MustSize32FromBytes(v)
}

/////////////////////////////////  Array  \\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\

// Array represents a dynamic array stored in a kv.KVStore
type Array struct {
	*ArrayReadOnly
	kvw kv.KVWriter
}

func NewArray(kvStore kv.KVStore, name string) *Array {
	return &Array{
		ArrayReadOnly: NewArrayReadOnly(kvStore, name),
		kvw:           kvStore,
	}
}

func (a *Array) addToSize(amount uint32) uint32 {
	oldSize := a.Len()
	newSize := oldSize + amount
	if newSize < oldSize {
		panic(ErrArrayOverflow)
	}
	a.setSize(newSize)
	return oldSize
}

// TODO implement with DelPrefix
func (a *Array) Erase() {
	length := a.Len()
	for i := uint32(0); i < length; i++ {
		a.kvw.Del(a.getArrayElemKey(i))
	}
	a.setSize(0)
}

func (a *Array) Extend(other *ArrayReadOnly) {
	length := other.Len()
	for i := uint32(0); i < length; i++ {
		a.Push(other.GetAt(i))
	}
}

func (a *Array) Immutable() *ArrayReadOnly {
	return a.ArrayReadOnly
}

// PruneAt deletes the value at the given index, without shifting the rest
// of the values.
func (a *Array) PruneAt(index uint32) {
	length := a.Len()
	if index >= length {
		panic(fmt.Errorf("index %d out of range for array of len %d", index, length))
	}
	a.kvw.Del(a.getArrayElemKey(index))
}

// adds to the end of the list
func (a *Array) Push(value []byte) {
	prevSize := a.addToSize(1)
	k := a.getArrayElemKey(prevSize)
	a.kvw.Set(k, value)
}

func (a *Array) SetAt(index uint32, value []byte) {
	length := a.Len()
	if index >= length {
		panic(fmt.Errorf("index %d out of range for array of len %d", index, length))
	}
	a.kvw.Set(a.getArrayElemKey(index), value)
}

func (a *Array) setSize(size uint32) {
	if size == 0 {
		a.kvw.Del(a.getSizeKey())
	} else {
		a.kvw.Set(a.getSizeKey(), rwutil.Size32ToBytes(size))
	}
}
