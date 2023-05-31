package collections

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
)

var ErrArray16Overflow = errors.New("Array16 overflow")

// For easy distinction between arrays and map collections
// we use '#' as separator for arrays and '.' for maps.
// Do not change this value unless you want to break how
// WasmLib maps these collections in the exact same way
const array16ElemKeyCode = byte('#')

func Array16ElemKey(name string, index uint16) kv.Key {
	var buf bytes.Buffer
	buf.Write([]byte(name))
	buf.WriteByte(array16ElemKeyCode)
	buf.Write(util.Size16ToBytes(index))
	return kv.Key(buf.Bytes())
}

/////////////////////////////////  Array16ReadOnly  \\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\

// Array16ReadOnly provides read-only access to an Array16 in a kv.KVStoreReader.
type Array16ReadOnly struct {
	kvr  kv.KVStoreReader
	name string
}

func NewArray16ReadOnly(kvReader kv.KVStoreReader, name string) *Array16ReadOnly {
	return &Array16ReadOnly{
		kvr:  kvReader,
		name: name,
	}
}

func (a *Array16ReadOnly) getArray16ElemKey(index uint16) kv.Key {
	return Array16ElemKey(a.name, index)
}

func (a *Array16ReadOnly) GetAt(index uint16) []byte {
	length := a.Len()
	if index >= length {
		panic(fmt.Errorf("index %d out of range for array of len %d", index, length))
	}
	return a.kvr.Get(a.getArray16ElemKey(index))
}

func (a *Array16ReadOnly) getSizeKey() kv.Key {
	return kv.Key(a.name)
}

// Len == 0/empty/non-existent are equivalent
func (a *Array16ReadOnly) Len() uint16 {
	v := a.kvr.Get(a.getSizeKey())
	if v == nil {
		return 0
	}
	return util.BytesToSize16(v)
}

/////////////////////////////////  Array16  \\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\

// Array16 represents a dynamic array stored in a kv.KVStore
type Array16 struct {
	*Array16ReadOnly
	kvw kv.KVWriter
}

func NewArray16(kvStore kv.KVStore, name string) *Array16 {
	return &Array16{
		Array16ReadOnly: NewArray16ReadOnly(kvStore, name),
		kvw:             kvStore,
	}
}

func (a *Array16) addToSize(amount uint16) uint16 {
	prevSize := a.Len()
	newSize := prevSize + amount
	if newSize < prevSize {
		panic(ErrArray16Overflow)
	}
	a.setSize(newSize)
	return prevSize
}

// TODO implement with DelPrefix
func (a *Array16) Erase() {
	length := a.Len()
	for i := uint16(0); i < length; i++ {
		a.kvw.Del(a.getArray16ElemKey(i))
	}
	a.setSize(0)
}

func (a *Array16) Extend(other *Array16ReadOnly) {
	length := other.Len()
	for i := uint16(0); i < length; i++ {
		a.Push(other.GetAt(i))
	}
}

func (a *Array16) Immutable() *Array16ReadOnly {
	return a.Array16ReadOnly
}

// PruneAt deletes the value at the given index, without shifting the rest
// of the values.
func (a *Array16) PruneAt(index uint16) {
	length := a.Len()
	if index >= length {
		panic(fmt.Errorf("index %d out of range for array of len %d", index, length))
	}
	a.kvw.Del(a.getArray16ElemKey(index))
}

// adds to the end of the list
func (a *Array16) Push(value []byte) {
	prevSize := a.addToSize(1)
	k := a.getArray16ElemKey(prevSize)
	a.kvw.Set(k, value)
}

func (a *Array16) SetAt(index uint16, value []byte) {
	length := a.Len()
	if index >= length {
		panic(fmt.Errorf("index %d out of range for array of len %d", index, length))
	}
	a.kvw.Set(a.getArray16ElemKey(index), value)
}

func (a *Array16) setSize(size uint16) {
	if size == 0 {
		a.kvw.Del(a.getSizeKey())
	} else {
		a.kvw.Set(a.getSizeKey(), util.Size16ToBytes(size))
	}
}
