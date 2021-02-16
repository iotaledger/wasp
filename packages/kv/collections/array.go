package collections

import (
	"bytes"
	"fmt"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
)

// Array represents a dynamic array stored in a kv.KVStore
type Array struct {
	*ImmutableArray
	kvw kv.KVStoreWriter
}

// ImmutableArray provides read-only access to an Array in a kv.KVStoreReader.
type ImmutableArray struct {
	kvr  kv.KVStoreReader
	name string
}

func NewArray(kv kv.KVStore, name string) *Array {
	return &Array{
		ImmutableArray: NewArrayReadOnly(kv, name),
		kvw:            kv,
	}
}

func NewArrayReadOnly(kv kv.KVStoreReader, name string) *ImmutableArray {
	return &ImmutableArray{
		kvr:  kv,
		name: name,
	}
}

const (
	arraySizeKeyCode = byte(0)
	arrayElemKeyCode = byte(1)
)

func (a *Array) Immutable() *ImmutableArray {
	return a.ImmutableArray
}

func (a *ImmutableArray) getSizeKey() kv.Key {
	return ArraySizeKey(a.name)
}

func ArraySizeKey(name string) kv.Key {
	var buf bytes.Buffer
	buf.Write([]byte(name))
	buf.WriteByte(arraySizeKeyCode)
	return kv.Key(buf.Bytes())
}

func (a *ImmutableArray) getElemKey(idx uint16) kv.Key {
	return ArrayElemKey(a.name, idx)
}

func ArrayElemKey(name string, idx uint16) kv.Key {
	var buf bytes.Buffer
	buf.Write([]byte(name))
	buf.WriteByte(arrayElemKeyCode)
	_ = util.WriteUint16(&buf, idx)
	return kv.Key(buf.Bytes())
}

// ArrayRangeKeys returns the KVStore keys for the items between [from, to) (`to` being not inclusive),
// assuming it has `length` elements.
func ArrayRangeKeys(name string, length uint16, from uint16, to uint16) []kv.Key {
	keys := make([]kv.Key, 0)
	if to >= from {
		for i := from; i < to && i < length; i++ {
			keys = append(keys, ArrayElemKey(name, i))
		}
	}
	return keys
}

func (a *Array) setSize(n uint16) {
	if n == 0 {
		a.kvw.Del(a.getSizeKey())
	} else {
		a.kvw.Set(a.getSizeKey(), util.Uint16To2Bytes(n))
	}
}

func (a *Array) addToSize(amount int) (uint16, error) {
	prevSize, err := a.Len()
	if err != nil {
		return 0, err
	}
	a.setSize(uint16(int(prevSize) + amount))
	return prevSize, nil
}

// Len == 0/empty/non-existent are equivalent
func (a *ImmutableArray) Len() (uint16, error) {
	v, err := a.kvr.Get(a.getSizeKey())
	if err != nil {
		return 0, err
	}
	if v == nil {
		return 0, nil
	}
	return util.MustUint16From2Bytes(v), nil
}

func (a *ImmutableArray) MustLen() uint16 {
	n, err := a.Len()
	if err != nil {
		panic(err)
	}
	return n
}

// adds to the end of the list
func (a *Array) Push(value []byte) error {
	prevSize, err := a.addToSize(1)
	if err != nil {
		return err
	}
	k := a.getElemKey(prevSize)
	a.kvw.Set(k, value)
	return nil
}

func (a *Array) MustPush(value []byte) {
	err := a.Push(value)
	if err != nil {
		panic(err)
	}
}

func (a *Array) Extend(other *ImmutableArray) error {
	otherLen, err := other.Len()
	if err != nil {
		return err
	}
	for i := uint16(0); i < otherLen; i++ {
		v, _ := other.GetAt(i)
		err = a.Push(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Array) MustExtend(other *ImmutableArray) {
	err := a.Extend(other)
	if err != nil {
		panic(err)
	}
}

// TODO implement with DelPrefix
func (a *Array) Erase() error {
	n, err := a.Len()
	if err != nil {
		return err
	}
	for i := uint16(0); i < n; i++ {
		a.kvw.Del(a.getElemKey(i))
	}
	a.setSize(0)
	return nil
}

func (a *Array) MustErase() {
	err := a.Erase()
	if err != nil {
		panic(err)
	}
}

func (a *ImmutableArray) GetAt(idx uint16) ([]byte, error) {
	n, err := a.Len()
	if err != nil {
		return nil, err
	}
	if idx >= n {
		return nil, fmt.Errorf("index %d out of range for array of len %d", idx, n)
	}
	ret, err := a.kvr.Get(a.getElemKey(idx))
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (a *ImmutableArray) MustGetAt(idx uint16) []byte {
	ret, err := a.GetAt(idx)
	if err != nil {
		panic(err)
	}
	return ret
}

func (a *Array) SetAt(idx uint16, value []byte) error {
	n, err := a.Len()
	if err != nil {
		return err
	}
	if idx >= n {
		return fmt.Errorf("index %d out of range for array of len %d", idx, n)
	}
	a.kvw.Set(a.getElemKey(idx), value)
	return nil
}

func (a *Array) MustSetAt(idx uint16, value []byte) {
	err := a.SetAt(idx, value)
	if err != nil {
		panic(err)
	}
}
