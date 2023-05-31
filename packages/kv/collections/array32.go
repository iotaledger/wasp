package collections

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
)

var ErrArray32Overflow = errors.New("Array32 overflow")

// For easy distinction between arrays and map collections
// we use '#' as separator for arrays and '.' for maps.
// Do not change this value unless you want to break how
// WasmLib maps these collections in the exact same way
const array32ElemKeyCode = byte('#')

func Array32ElemKey(name string, index uint32) kv.Key {
	var buf bytes.Buffer
	buf.Write([]byte(name))
	buf.WriteByte(array32ElemKeyCode)
	buf.Write(util.Size32ToBytes(index))
	return kv.Key(buf.Bytes())
}

/////////////////////////////////  Array32ReadOnly  \\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\

// Array32ReadOnly provides read-only access to an Array32 in a kv.KVStoreReader.
type Array32ReadOnly struct {
	kvr  kv.KVStoreReader
	name string
}

func NewArray32ReadOnly(kvReader kv.KVStoreReader, name string) *Array32ReadOnly {
	return &Array32ReadOnly{
		kvr:  kvReader,
		name: name,
	}
}

func (a *Array32ReadOnly) getArray32ElemKey(index uint32) kv.Key {
	return Array32ElemKey(a.name, index)
}

func (a *Array32ReadOnly) GetAt(index uint32) []byte {
	length := a.Len()
	if index >= length {
		panic(fmt.Errorf("index %d out of range for array of len %d", index, length))
	}
	return a.kvr.Get(a.getArray32ElemKey(index))
}

func (a *Array32ReadOnly) getSizeKey() kv.Key {
	return kv.Key(a.name)
}

// Len == 0/empty/non-existent are equivalent
func (a *Array32ReadOnly) Len() uint32 {
	v := a.kvr.Get(a.getSizeKey())
	if v == nil {
		return 0
	}
	return util.BytesToSize32(v)
}

/////////////////////////////////  Array32  \\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\

// Array32 represents a dynamic array stored in a kv.KVStore
type Array32 struct {
	*Array32ReadOnly
	kvw kv.KVWriter
}

func NewArray32(kvStore kv.KVStore, name string) *Array32 {
	return &Array32{
		Array32ReadOnly: NewArray32ReadOnly(kvStore, name),
		kvw:             kvStore,
	}
}

func (a *Array32) addToSize(amount uint32) uint32 {
	prevSize := a.Len()
	newSize := prevSize + amount
	if newSize < prevSize {
		panic(ErrArray32Overflow)
	}
	a.setSize(newSize)
	return prevSize
}

// TODO implement with DelPrefix
func (a *Array32) Erase() {
	length := a.Len()
	for i := uint32(0); i < length; i++ {
		a.kvw.Del(a.getArray32ElemKey(i))
	}
	a.setSize(0)
}

func (a *Array32) Extend(other *Array32ReadOnly) {
	length := other.Len()
	for i := uint32(0); i < length; i++ {
		a.Push(other.GetAt(i))
	}
}

func (a *Array32) Immutable() *Array32ReadOnly {
	return a.Array32ReadOnly
}

// PruneAt deletes the value at the given index, without shifting the rest
// of the values.
func (a *Array32) PruneAt(index uint32) {
	length := a.Len()
	if index >= length {
		panic(fmt.Errorf("index %d out of range for array of len %d", index, length))
	}
	a.kvw.Del(a.getArray32ElemKey(index))
}

// adds to the end of the list
func (a *Array32) Push(value []byte) {
	prevSize := a.addToSize(1)
	k := a.getArray32ElemKey(prevSize)
	a.kvw.Set(k, value)
}

func (a *Array32) SetAt(index uint32, value []byte) {
	length := a.Len()
	if index >= length {
		panic(fmt.Errorf("index %d out of range for array of len %d", index, length))
	}
	a.kvw.Set(a.getArray32ElemKey(index), value)
}

func (a *Array32) setSize(size uint32) {
	if size == 0 {
		a.kvw.Del(a.getSizeKey())
	} else {
		a.kvw.Set(a.getSizeKey(), util.Size32ToBytes(size))
	}
}
