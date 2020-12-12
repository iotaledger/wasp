package datatypes

import (
	"bytes"
	"errors"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
)

type Map struct {
	kv   kv.KVStore
	name string
}

type MustMap struct {
	m Map
}

const (
	mapSizeKeyCode = byte(0)
	mapElemKeyCode = byte(1)
)

func NewMap(kv kv.KVStore, name string) *Map {
	return &Map{
		kv:   kv,
		name: name,
	}
}

func NewMustMap(kv kv.KVStore, name string) *MustMap {
	return NewMap(kv, name).Must()
}

func (m *Map) Must() *MustMap {
	return &MustMap{*m}
}

func (m *Map) Name() string {
	return m.name
}

func (m *MustMap) Name() string {
	return m.m.name
}

func (m *Map) getSizeKey() kv.Key {
	var buf bytes.Buffer
	buf.Write([]byte(m.name))
	buf.WriteByte(mapSizeKeyCode)
	return kv.Key(buf.Bytes())
}

func (m *Map) getElemKey(key []byte) kv.Key {
	var buf bytes.Buffer
	buf.Write([]byte(m.name))
	buf.WriteByte(mapElemKeyCode)
	buf.Write(key)
	return kv.Key(buf.Bytes())
}

func (m *Map) addToSize(amount int) error {
	n, err := m.Len()
	if err != nil {
		return err
	}
	n = uint32(int(n) + amount)
	if n == 0 {
		m.kv.Del(m.getSizeKey())
	} else {
		m.kv.Set(m.getSizeKey(), util.Uint32To4Bytes(n))
	}
	return nil
}

func (d *Map) GetAt(key []byte) ([]byte, error) {
	ret, err := d.kv.Get(d.getElemKey(key))
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (d *MustMap) GetAt(key []byte) []byte {
	ret, err := d.m.GetAt(key)
	if err != nil {
		panic(err)
	}
	return ret
}

func (d *Map) SetAt(key []byte, value []byte) error {
	ok, err := d.HasAt(key)
	if err != nil {
		return err
	}
	if !ok {
		err = d.addToSize(1)
		if err != nil {
			return err
		}
	}
	d.kv.Set(d.getElemKey(key), value)
	return nil
}

func (d *MustMap) SetAt(key []byte, value []byte) {
	_ = d.m.SetAt(key, value)
}

func (d *Map) DelAt(key []byte) error {
	ok, err := d.HasAt(key)
	if err != nil {
		return err
	}
	if ok {
		err = d.addToSize(-1)
		if err != nil {
			return err
		}
	}
	d.kv.Del(d.getElemKey(key))
	return nil
}

func (d *MustMap) DelAt(key []byte) {
	_ = d.m.DelAt(key)
}

func (d *Map) HasAt(key []byte) (bool, error) {
	return d.kv.Has(d.getElemKey(key))
}

func (d *MustMap) HasAt(key []byte) bool {
	ret, err := d.m.HasAt(key)
	if err != nil {
		panic(err)
	}
	return ret
}

func (d *MustMap) Len() uint32 {
	n, err := d.m.Len()
	if err != nil {
		panic(err)
	}
	return n
}

func (d *Map) Len() (uint32, error) {
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
	return util.MustUint32From4Bytes(v), nil
}

func (d *Map) Erase() {
	// TODO needs DelPrefix method in KVStore
	panic("implement me")
}

// Iterate non-deterministic
func (d *Map) Iterate(f func(elemKey []byte, value []byte) bool) error {
	prefix := d.getElemKey(nil)
	return d.kv.Iterate(prefix, func(key kv.Key, value []byte) bool {
		return f([]byte(key)[len(prefix):], value)
		//return f([]byte(key), value)
	})
}

// Iterate non-deterministic
func (d *Map) IterateKeys(f func(elemKey []byte) bool) error {
	prefix := d.getElemKey(nil)
	return d.kv.IterateKeys(prefix, func(key kv.Key) bool {
		return f([]byte(key)[len(prefix):])
	})
}

// Iterate non-deterministic
func (d *MustMap) Iterate(f func(elemKey []byte, value []byte) bool) {
	err := d.m.Iterate(f)
	if err != nil {
		panic(err)
	}
}

// Iterate non-deterministic
func (d *MustMap) IterateKeys(f func(elemKey []byte) bool) {
	err := d.m.IterateKeys(f)
	if err != nil {
		panic(err)
	}
}

func (d *MustMap) IterateBalances(f func(color balance.Color, bal int64) bool) error {
	var err error
	d.Iterate(func(elemKey []byte, value []byte) bool {
		col, _, err := balance.ColorFromBytes(elemKey)
		if err != nil {
			return false
		}
		v, err := util.Uint64From8Bytes(value)
		if err != nil {
			return false
		}
		bal := int64(v)
		return f(col, bal)
	})
	return err
}
