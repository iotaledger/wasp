package collections

import (
	"bytes"
	"errors"
	"fmt"
	"math"

	"github.com/iotaledger/wasp/packages/kv"
)

var ErrArray16Overflow = errors.New("Array16 overflow")

// Array16 represents a dynamic array stored in a kv.KVStore
type Array16 struct {
	*ImmutableArray16
	kvw kv.KVWriter
}

// ImmutableArray16 provides read-only access to an Array16 in a kv.KVStoreReader.
type ImmutableArray16 struct {
	kvr  kv.KVStoreReader
	name string
}

func NewArray16(kvStore kv.KVStore, name string) *Array16 {
	return &Array16{
		ImmutableArray16: NewArray16ReadOnly(kvStore, name),
		kvw:              kvStore,
	}
}

func NewArray16ReadOnly(kvReader kv.KVStoreReader, name string) *ImmutableArray16 {
	return &ImmutableArray16{
		kvr:  kvReader,
		name: name,
	}
}

// For easy distinction between arrays and map collections
// we use '#' as separator for arrays and '.' for maps.
// Do not change this value unless you want to break how
// WasmLib maps these collections in the exact same way
const array16ElemKeyCode = byte('#')

func array16SizeKey(name string) kv.Key {
	return kv.Key(name)
}

func array16ElemKey(name string, idx uint16) kv.Key {
	var buf bytes.Buffer
	buf.Write([]byte(name))
	buf.WriteByte(array16ElemKeyCode)
	buf.Write(uint16ToBytes(idx))
	return kv.Key(buf.Bytes())
}

// Array16RangeKeys returns the KVStore keys for the items between [from, to) (`to` being not inclusive),
// assuming it has `length` elements.
func Array16RangeKeys(name string, length, from, to uint16) []kv.Key {
	keys := make([]kv.Key, 0)
	if to >= from {
		for i := from; i < to && i < length; i++ {
			keys = append(keys, array16ElemKey(name, i))
		}
	}
	return keys
}

// use ULEB128 decoding so that WasmLib can use it as well
func bytesToUint16(buf []byte) uint16 {
	if (buf[0] & 0x80) == 0 {
		return uint16(buf[0])
	}
	if (buf[1] & 0x80) == 0 {
		return (uint16(buf[1]) << 7) | uint16(buf[0]&0x7f)
	}
	return (uint16(buf[2]) << 14) | (uint16(buf[1]&0x7f) << 7) | uint16(buf[0]&0x7f)
}

// use ULEB128 encoding so that WasmLib can decode it as well
func uint16ToBytes(value uint16) []byte {
	if value < 128 {
		return []byte{byte(value)}
	}
	if value < 16384 {
		return []byte{byte(value | 0x80), byte(value >> 7)}
	}
	return []byte{byte(value | 0x80), byte((value >> 7) | 0x80), byte(value >> 14)}
}

func (a *Array16) Immutable() *ImmutableArray16 {
	return a.ImmutableArray16
}

func (a *ImmutableArray16) getSizeKey() kv.Key {
	return array16SizeKey(a.name)
}

func (a *ImmutableArray16) getArray16ElemKey(idx uint16) kv.Key {
	return array16ElemKey(a.name, idx)
}

func (a *Array16) setSize(n uint16) {
	if n == 0 {
		a.kvw.Del(a.getSizeKey())
	} else {
		a.kvw.Set(a.getSizeKey(), uint16ToBytes(n))
	}
}

func (a *Array16) addToSize(amount int) uint16 {
	prevSize := a.Len()
	newSize := int(prevSize) + amount
	if newSize > math.MaxUint16 {
		panic(ErrArray16Overflow)
	}
	a.setSize(uint16(newSize))
	return prevSize
}

// Len == 0/empty/non-existent are equivalent
func (a *ImmutableArray16) Len() uint16 {
	v := a.kvr.Get(a.getSizeKey())
	if v == nil {
		return 0
	}
	return bytesToUint16(v)
}

// adds to the end of the list
func (a *Array16) Push(value []byte) {
	prevSize := a.addToSize(1)
	k := a.getArray16ElemKey(prevSize)
	a.kvw.Set(k, value)
}

func (a *Array16) Extend(other *ImmutableArray16) {
	otherLen := other.Len()
	for i := uint16(0); i < otherLen; i++ {
		a.Push(other.GetAt(i))
	}
}

// TODO implement with DelPrefix
func (a *Array16) Erase() {
	n := a.Len()
	for i := uint16(0); i < n; i++ {
		a.kvw.Del(a.getArray16ElemKey(i))
	}
	a.setSize(0)
}

func (a *ImmutableArray16) GetAt(idx uint16) []byte {
	n := a.Len()
	if idx >= n {
		panic(fmt.Errorf("index %d out of range for array of len %d", idx, n))
	}
	return a.kvr.Get(a.getArray16ElemKey(idx))
}

func (a *Array16) SetAt(idx uint16, value []byte) {
	n := a.Len()
	if idx >= n {
		panic(fmt.Errorf("index %d out of range for array of len %d", idx, n))
	}
	a.kvw.Set(a.getArray16ElemKey(idx), value)
}
