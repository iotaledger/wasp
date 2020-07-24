package kv

import (
	"bytes"
	"errors"

	"github.com/iotaledger/wasp/packages/util"
)

type Dictionary struct {
	kv         KVStore
	name       string
	cachedsize uint32
}

type MustDictionary struct {
	dict Dictionary
}

const (
	dictSizeKeyCode = byte(0)
	dictElemKeyCode = byte(1)
)

func newDictionary(kv KVStore, name string) (*Dictionary, error) {
	ret := &Dictionary{
		kv:   kv,
		name: name,
	}
	var err error
	ret.cachedsize, err = ret.len()
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func newMustDictionary(dict *Dictionary) *MustDictionary {
	return &MustDictionary{*dict}
}

func (l *Dictionary) getSizeKey() Key {
	var buf bytes.Buffer
	buf.Write([]byte(l.name))
	buf.WriteByte(dictSizeKeyCode)
	return Key(buf.Bytes())
}

func (l *Dictionary) getElemKey(key []byte) Key {
	var buf bytes.Buffer
	buf.Write([]byte(l.name))
	buf.WriteByte(dictElemKeyCode)
	buf.Write(key)
	return Key(buf.Bytes())
}

func (l *Dictionary) setSize(size uint32) {
	if size == 0 {
		l.kv.Del(l.getSizeKey())
		return
	}
	l.cachedsize = size
	l.kv.Set(l.getSizeKey(), util.Uint32To4Bytes(size))
}

func (d *Dictionary) GetAt(key []byte) ([]byte, error) {
	ret, err := d.kv.Get(d.getElemKey(key))
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (d *MustDictionary) GetAt(key []byte) []byte {
	ret, err := d.dict.GetAt(key)
	if err != nil {
		panic(err)
	}
	return ret
}

func (d *Dictionary) SetAt(key []byte, value []byte) error {
	if d.Len() == 0 {
		d.setSize(1)
	} else {
		ok, err := d.HasAt(key)
		if err != nil {
			return err
		}
		if !ok {
			d.setSize(d.Len() + 1)
		}
	}
	d.kv.Set(d.getElemKey(key), value)
	return nil
}

func (d *MustDictionary) SetAt(key []byte, value []byte) {
	_ = d.dict.SetAt(key, value)
}

func (d *Dictionary) DelAt(key []byte) error {
	ok, err := d.HasAt(key)
	if err != nil {
		return err
	}
	if ok {
		d.setSize(d.Len() - 1)
	}
	d.kv.Del(d.getElemKey(key))
	return nil
}

func (d *MustDictionary) DelAt(key []byte) {
	_ = d.dict.DelAt(key)
}

func (d *Dictionary) HasAt(key []byte) (bool, error) {
	return d.kv.Has(d.getElemKey(key))
}

func (d *Dictionary) Len() uint32 {
	return d.cachedsize
}

func (d *Dictionary) len() (uint32, error) {
	v, err := d.kv.Get(d.getSizeKey())
	if err != nil {
		return 0, err
	}
	if v == nil {
		return 0, nil
	}
	if len(v) != 4 {
		return 0, errors.New("corrupted data")
	}
	return util.Uint32From4Bytes(v), nil
}

func (d *Dictionary) Erase() {
	// TODO needs DelPrefix method in KVStore
	panic("implement me")
}

func (d *Dictionary) Iterate(f func(elemKey []byte, value []byte) bool) error {
	prefix := d.getElemKey([]byte{})
	return d.kv.Iterate(prefix, func(key Key, value []byte) bool {
		return f([]byte(key[len(prefix):]), value)
	})
}
