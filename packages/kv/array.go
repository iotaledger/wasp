package kv

import (
	"bytes"
	"errors"
	"github.com/iotaledger/wasp/packages/util"
)

type Array struct {
	kv        KVStore
	name      string
	cachedLen uint16
}

func NewArray(kv KVStore, name string) (*Array, error) {
	ret := &Array{
		kv:   kv,
		name: name,
	}
	var err error
	ret.cachedLen, err = ret.len()
	if err != nil {
		return nil, err
	}
	return ret, nil
}

const (
	arraySizeKeyCode = byte(0)
	arrayElemKeyCode = byte(1)
)

func (l *Array) getSizeKey() Key {
	return ArraySizeKey(l.name)
}

func ArraySizeKey(name string) Key {
	var buf bytes.Buffer
	buf.Write([]byte(name))
	buf.WriteByte(arraySizeKeyCode)
	return Key(buf.Bytes())
}

func (l *Array) getElemKey(idx uint16) Key {
	return ArrayElemKey(l.name, idx)
}

func ArrayElemKey(name string, idx uint16) Key {
	var buf bytes.Buffer
	buf.Write([]byte(name))
	buf.WriteByte(arrayElemKeyCode)
	_ = util.WriteUint16(&buf, idx)
	return Key(buf.Bytes())
}

// ArrayRangeKeys returns the KVStore keys for the items between [from, to) (`to` being not inclusive),
// assuming it has `length` elements.
func ArrayRangeKeys(name string, length uint16, from uint16, to uint16) []Key {
	keys := make([]Key, 0)
	if to >= from {
		for i := from; i < to && i < length; i++ {
			keys = append(keys, ArrayElemKey(name, i))
		}
	}
	return keys
}

func (l *Array) setSize(size uint16) {
	if size == 0 {
		l.kv.Del(l.getSizeKey())
		l.cachedLen = 0
		return
	}
	l.cachedLen = size
	l.kv.Set(l.getSizeKey(), util.Uint16To2Bytes(size))
}

// Len == 0/empty/non-existent are equivalent
func (l *Array) Len() uint16 {
	return l.cachedLen
}

func (l *Array) len() (uint16, error) {
	v, err := l.kv.Get(l.getSizeKey())
	if err != nil {
		return 0, err
	}
	if v == nil {
		return 0, nil
	}
	if len(v) != 2 {
		return 0, errors.New("corrupted data")
	}
	return util.Uint16From2Bytes(v), nil
}

// adds to the end of the list
func (l *Array) Push(value []byte) {
	size := l.Len()
	k := l.getElemKey(size)
	l.kv.Set(k, value)
	l.setSize(size + 1)
}

func (l *Array) Append(arr *Array) {
	for i := uint16(0); i < arr.Len(); i++ {
		v, _ := arr.GetAt(i)
		l.Push(v)
	}
}

// TODO implement with DelPrefix
func (l *Array) Erase() {
	for i := uint16(0); i < l.Len(); i++ {
		l.kv.Del(l.getElemKey(i))
	}
	l.setSize(0)
}

func (l *Array) GetAt(idx uint16) ([]byte, error) {
	if idx >= l.Len() {
		return nil, errors.New("index out of range")
	}
	ret, err := l.kv.Get(l.getElemKey(idx))
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (l *Array) MustGetAt(idx uint16) []byte {
	ret, err := l.GetAt(idx)
	if err != nil {
		panic(err)
	}
	return ret
}

func (l *Array) SetAt(idx uint16, value []byte) bool {
	if idx >= l.Len() {
		return false
	}
	l.kv.Set(l.getElemKey(idx), value)
	return true
}
