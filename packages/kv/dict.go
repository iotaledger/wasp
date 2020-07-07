package kv

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/util"
)

// TODO iteration over dictionary ??

type Dictionary interface {
	GetAt(key []byte) ([]byte, bool)
	SetAt(key []byte, value []byte)
	DelAt(key []byte)
	ExistsAt(key []byte) bool
	Len() uint32
	Erase()
}

type dictStruct struct {
	kv   KVStore
	name string
}

const (
	dictSizeKeyCode = byte(0)
	dictElemKeyCode = byte(1)
)

func newDict(kv KVStore, name string) Dictionary {
	return &dictStruct{
		kv:   kv,
		name: name,
	}
}

func (l *dictStruct) getSizeKey() Key {
	var buf bytes.Buffer
	buf.Write([]byte(l.name))
	buf.WriteByte(dictSizeKeyCode)
	return Key(buf.Bytes())
}

func (l *dictStruct) getElemKey(key []byte) Key {
	var buf bytes.Buffer
	buf.Write([]byte(l.name))
	buf.WriteByte(dictElemKeyCode)
	_, _ = buf.Write(key)
	return Key(buf.Bytes())
}

func (l *dictStruct) setSize(size uint32) {
	if size == 0 {
		l.kv.Del(l.getSizeKey())
		return
	}
	l.kv.Set(l.getSizeKey(), util.Uint32To4Bytes(size))
}

func (d *dictStruct) GetAt(key []byte) ([]byte, bool) {
	ret, err := d.kv.Get(d.getElemKey(key))
	return ret, err != nil
}

func (d *dictStruct) SetAt(key []byte, value []byte) {
	if !d.ExistsAt(key) {
		d.setSize(d.Len() + 1)
	}
	d.kv.Set(d.getElemKey(key), value)
}

func (d *dictStruct) DelAt(key []byte) {
	if d.ExistsAt(key) {
		d.setSize(d.Len() - 1)
	}
	d.kv.Del(d.getElemKey(key))
}

func (d *dictStruct) ExistsAt(key []byte) bool {
	// TODO implement with Has
	_, err := d.kv.Get(d.getElemKey(key))
	return err == nil
}

func (d *dictStruct) Len() uint32 {
	v, err := d.kv.Get(d.getSizeKey())
	if err != nil || len(v) != 4 {
		return 0
	}
	return util.Uint32From4Bytes(v)
}

func (d *dictStruct) Erase() {
	// TODO needs iteration
	panic("implement me")
}
