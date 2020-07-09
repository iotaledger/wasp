package kv

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/util"
)

type Array interface {
	Push(value []byte)
	Append(Array)
	Len() uint16
	Erase()
	GetAt(uint16) ([]byte, bool)
	SetAt(idx uint16, value []byte) bool
}

type arrayStruct struct {
	kv   KVStore
	name string
}

func newArray(kv KVStore, name string) Array {
	return &arrayStruct{
		kv:   kv,
		name: name,
	}
}

const (
	arraySizeKeyCode = byte(0)
	arrayElemKeyCode = byte(1)
)

func (l *arrayStruct) getSizeKey() Key {
	return ArraySizeKey(l.name)
}

func ArraySizeKey(name string) Key {
	var buf bytes.Buffer
	buf.Write([]byte(name))
	buf.WriteByte(arraySizeKeyCode)
	return Key(buf.Bytes())
}

func (l *arrayStruct) getElemKey(idx uint16) Key {
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
		for i := uint16(from); i < to && i < length; i++ {
			keys = append(keys, ArrayElemKey(name, i))
		}
	}
	return keys
}

func (l *arrayStruct) setSize(size uint16) {
	if size == 0 {
		l.kv.Del(l.getSizeKey())
		return
	}
	l.kv.Set(l.getSizeKey(), util.Uint16To2Bytes(size))
}

// Len == 0/empty/non-existent are equivalent
func (l *arrayStruct) Len() uint16 {
	v, err := l.kv.Get(l.getSizeKey())
	if err != nil || len(v) != 2 {
		return 0
	}
	return util.Uint16From2Bytes(v)
}

// adds to the end of the list
func (l *arrayStruct) Push(value []byte) {
	size := l.Len()
	k := l.getElemKey(size)
	l.kv.Set(k, value)
	l.setSize(size + 1)
}

func (l *arrayStruct) Append(arr Array) {
	for i := uint16(0); i < arr.Len(); i++ {
		v, _ := arr.GetAt(i)
		l.Push(v)
	}
}

func (l *arrayStruct) Erase() {
	for i := uint16(0); i < l.Len(); i++ {
		l.kv.Del(l.getElemKey(i))
	}
	l.setSize(0)
}

func (l *arrayStruct) GetAt(idx uint16) ([]byte, bool) {
	if idx >= l.Len() {
		return nil, false
	}
	ret, err := l.kv.Get(l.getElemKey(idx))
	if err != nil {
		return nil, false
	}
	return ret, true
}

func (l *arrayStruct) SetAt(idx uint16, value []byte) bool {
	if idx >= l.Len() {
		return false
	}
	l.kv.Set(l.getElemKey(idx), value)
	return true
}
