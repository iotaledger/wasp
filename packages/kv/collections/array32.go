package collections

import (
	"bytes"
	"errors"
	"fmt"
	"math"

	"github.com/iotaledger/wasp/packages/kv"
)

var ErrArray32Overflow = errors.New("Array32 overflow")

// Array32 represents a dynamic array stored in a kv.KVStore
type Array32 struct {
	*ImmutableArray32
	kvw kv.KVWriter
}

// ImmutableArray32 provides read-only access to an Array32 in a kv.KVStoreReader.
type ImmutableArray32 struct {
	kvr  kv.KVStoreReader
	name string
}

func NewArray32(kvStore kv.KVStore, name string) *Array32 {
	return &Array32{
		ImmutableArray32: NewArray32ReadOnly(kvStore, name),
		kvw:              kvStore,
	}
}

func NewArray32ReadOnly(kvReader kv.KVStoreReader, name string) *ImmutableArray32 {
	return &ImmutableArray32{
		kvr:  kvReader,
		name: name,
	}
}

// For easy distinction between arrays and map collections
// we use '#' as separator for arrays and '.' for maps.
// Do not change this value unless you want to break how
// WasmLib maps these collections in the exact same way
const array32ElemKeyCode = byte('#')

func Array32SizeKey(name string) kv.Key {
	return kv.Key(name)
}

func Array32ElemKey(name string, idx uint32) kv.Key {
	var buf bytes.Buffer
	buf.Write([]byte(name))
	buf.WriteByte(array32ElemKeyCode)
	buf.Write(uint32ToBytes(idx))
	return kv.Key(buf.Bytes())
}

// Array32RangeKeys returns the KVStore keys for the items between [from, to) (`to` being not inclusive),
// assuming it has `length` elements.
func Array32RangeKeys(name string, length, from, to uint32) []kv.Key {
	keys := make([]kv.Key, 0)
	if to >= from {
		for i := from; i < to && i < length; i++ {
			keys = append(keys, Array32ElemKey(name, i))
		}
	}
	return keys
}

// use ULEB128 decoding so that WasmLib can use it as well
func bytesToUint32(buf []byte) uint32 {
	value := uint32(buf[0] & 0x7f)
	i := 0
	for s := 7; (buf[i] & 0x80) != 0; s += 7 {
		i++
		value |= uint32(buf[i]&0x7f) << s
	}
	return value
}

// use ULEB128 encoding so that WasmLib can decode it as well
func uint32ToBytes(value uint32) []byte {
	buf := make([]byte, 0, 5)
	b := byte(value)
	value >>= 7
	for value != 0 {
		buf = append(buf, b|0x80)
		b = byte(value)
		value >>= 7
	}
	buf = append(buf, b)
	return buf
}

func (a *Array32) Immutable() *ImmutableArray32 {
	return a.ImmutableArray32
}

func (a *ImmutableArray32) getSizeKey() kv.Key {
	return Array32SizeKey(a.name)
}

func (a *ImmutableArray32) getArray32ElemKey(idx uint32) kv.Key {
	return Array32ElemKey(a.name, idx)
}

func (a *Array32) setSize(n uint32) {
	if n == 0 {
		a.kvw.Del(a.getSizeKey())
	} else {
		a.kvw.Set(a.getSizeKey(), uint32ToBytes(n))
	}
}

func (a *Array32) addToSize(amount int) uint32 {
	prevSize := a.Len()
	newSize := uint64(prevSize) + uint64(amount)
	if newSize > math.MaxUint32 {
		panic(ErrArray32Overflow)
	}
	a.setSize(uint32(newSize))
	return prevSize
}

// Len == 0/empty/non-existent are equivalent
func (a *ImmutableArray32) Len() uint32 {
	v := a.kvr.Get(a.getSizeKey())
	if v == nil {
		return 0
	}
	return bytesToUint32(v)
}

// adds to the end of the list
func (a *Array32) Push(value []byte) {
	prevSize := a.addToSize(1)
	k := a.getArray32ElemKey(prevSize)
	a.kvw.Set(k, value)
}

func (a *Array32) Extend(other *ImmutableArray32) {
	otherLen := other.Len()
	for i := uint32(0); i < otherLen; i++ {
		a.Push(other.GetAt(i))
	}
}

// TODO implement with DelPrefix
func (a *Array32) Erase() {
	n := a.Len()
	for i := uint32(0); i < n; i++ {
		a.kvw.Del(a.getArray32ElemKey(i))
	}
	a.setSize(0)
}

func (a *ImmutableArray32) GetAt(idx uint32) []byte {
	n := a.Len()
	if idx >= n {
		panic(fmt.Errorf("index %d out of range for array of len %d", idx, n))
	}
	return a.kvr.Get(a.getArray32ElemKey(idx))
}

func (a *Array32) SetAt(idx uint32, value []byte) {
	n := a.Len()
	if idx >= n {
		panic(fmt.Errorf("index %d out of range for array of len %d", idx, n))
	}
	a.kvw.Set(a.getArray32ElemKey(idx), value)
}
