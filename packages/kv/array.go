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
	At(uint16) ([]byte, bool)
}

type listStruct struct {
	kv   KVStore
	name string
}

func newListCodec(kv KVStore, name string) Array {
	return &listStruct{
		kv:   kv,
		name: name,
	}
}

const (
	sizeKeyCode = byte(0)
	elemKeyCode = byte(1)
)

func (l *listStruct) getSizeKey() Key {
	var buf bytes.Buffer
	buf.Write([]byte(l.name))
	buf.WriteByte(sizeKeyCode)
	return Key(buf.Bytes())
}

func (l *listStruct) getElemKey(idx uint16) Key {
	var buf bytes.Buffer
	buf.Write([]byte(l.name))
	buf.WriteByte(elemKeyCode)
	_ = util.WriteUint16(&buf, idx)
	return Key(buf.Bytes())
}

func (l *listStruct) setSize(size uint16) {
	if size == 0 {
		l.kv.Del(l.getSizeKey())
		return
	}
	l.kv.Set(l.getSizeKey(), util.Uint16To2Bytes(size))
}

// Len == 0/empty/non-existent are equivalent
func (l *listStruct) Len() uint16 {
	v, err := l.kv.Get(l.getSizeKey())
	if err != nil || len(v) != 2 {
		return 0
	}
	return util.Uint16From2Bytes(v)
}

// adds to the end of the list
func (l *listStruct) Push(value []byte) {
	size := l.Len()
	k := l.getElemKey(size)
	l.kv.Set(k, value)
	l.setSize(size + 1)
}

func (l *listStruct) Append(arr Array) {
	for i := uint16(0); i < arr.Len(); i++ {
		v, _ := arr.At(i)
		l.Push(v)
	}
}

func (l *listStruct) Erase() {
	for i := uint16(0); i < l.Len(); i++ {
		l.kv.Del(l.getElemKey(i))
	}
	l.setSize(0)
}

func (l *listStruct) At(idx uint16) ([]byte, bool) {
	if idx >= l.Len() {
		return nil, false
	}
	ret, err := l.kv.Get(l.getElemKey(idx))
	if err != nil {
		return nil, false
	}
	return ret, true
}
