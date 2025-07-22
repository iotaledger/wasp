package collections

import (
	"errors"
	"fmt"

	"fortio.org/safecast"

	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/util/rwutil"
)

var ErrArrayOverflow = errors.New("Array overflow")

// For easy distinction between arrays and map collections
// we use '#' as separator for arrays and '.' for maps.
const arrayElemKeyCode = byte('#')

func ArrayElemPrefix(name string) kv.Key {
	return kv.Key(name) + kv.Key([]byte{arrayElemKeyCode})
}

func ArrayElemKey(name string, index uint32) kv.Key {
	return kv.Key(rwutil.NewBytesWriter().
		WriteN([]byte(name)).
		WriteByte(arrayElemKeyCode).
		WriteSize32(int(index)).
		Bytes())
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

// Len == 0/empty/non-existent are equivalent
func (a *ArrayReadOnly) Len() uint32 {
	data := a.kvr.Get(kv.Key(a.name))
	if data == nil {
		return 0
	}
	size := rwutil.NewBytesReader(data).Must().ReadSize32()
	return safecast.MustConvert[uint32](size)
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

// Push adds to the end of the list
func (a *Array) Push(value []byte) {
	index := a.addToSize(1)
	key := a.getArrayElemKey(index)
	a.kvw.Set(key, value)
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
		a.kvw.Del(kv.Key(a.name))
		return
	}
	a.kvw.Set(kv.Key(a.name), rwutil.NewBytesWriter().WriteSize32(int(size)).Bytes())
}
