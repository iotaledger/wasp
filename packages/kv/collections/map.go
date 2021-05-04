package collections

import (
	"bytes"
	"errors"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
)

// Map represents a dynamic key-value collection in a kv.KVStore.
type Map struct {
	*ImmutableMap
	kvw kv.KVWriter
}

// ImmutableMap provides read-only access to a Map in a kv.KVStoreReader.
type ImmutableMap struct {
	kvr  kv.KVStoreReader
	name string
}

const (
	mapSizeKeyCode = byte(0)
	mapElemKeyCode = byte(1)
)

func NewMap(kv kv.KVStore, name string) *Map {
	return &Map{
		ImmutableMap: NewMapReadOnly(kv, name),
		kvw:          kv,
	}
}

func NewMapReadOnly(kv kv.KVStoreReader, name string) *ImmutableMap {
	return &ImmutableMap{
		kvr:  kv,
		name: name,
	}
}

func (m *Map) Immutable() *ImmutableMap {
	return m.ImmutableMap
}

func (m *ImmutableMap) Name() string {
	return m.name
}

func (m *ImmutableMap) getSizeKey() kv.Key {
	var buf bytes.Buffer
	buf.Write([]byte(m.name))
	buf.WriteByte(mapSizeKeyCode)
	return kv.Key(buf.Bytes())
}

func (m *ImmutableMap) getElemKey(key []byte) kv.Key {
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
		m.kvw.Del(m.getSizeKey())
	} else {
		m.kvw.Set(m.getSizeKey(), util.Uint32To4Bytes(n))
	}
	return nil
}

func (m *ImmutableMap) GetAt(key []byte) ([]byte, error) {
	ret, err := m.kvr.Get(m.getElemKey(key))
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (m *ImmutableMap) MustGetAt(key []byte) []byte {
	ret, err := m.GetAt(key)
	if err != nil {
		panic(err)
	}
	return ret
}

func (m *Map) SetAt(key []byte, value []byte) error {
	ok, err := m.HasAt(key)
	if err != nil {
		return err
	}
	if !ok {
		err = m.addToSize(1)
		if err != nil {
			return err
		}
	}
	m.kvw.Set(m.getElemKey(key), value)
	return nil
}

func (m *Map) MustSetAt(key []byte, value []byte) {
	err := m.SetAt(key, value)
	if err != nil {
		panic(err)
	}
}

func (m *Map) DelAt(key []byte) error {
	ok, err := m.HasAt(key)
	if err != nil {
		return err
	}
	if ok {
		err = m.addToSize(-1)
		if err != nil {
			return err
		}
	}
	m.kvw.Del(m.getElemKey(key))
	return nil
}

func (m *Map) MustDelAt(key []byte) {
	err := m.DelAt(key)
	if err != nil {
		panic(err)
	}
}

func (m *ImmutableMap) HasAt(key []byte) (bool, error) {
	return m.kvr.Has(m.getElemKey(key))
}

func (m *ImmutableMap) MustHasAt(key []byte) bool {
	ret, err := m.HasAt(key)
	if err != nil {
		panic(err)
	}
	return ret
}

func (m *ImmutableMap) MustLen() uint32 {
	n, err := m.Len()
	if err != nil {
		panic(err)
	}
	return n
}

func (m *ImmutableMap) Len() (uint32, error) {
	v, err := m.kvr.Get(m.getSizeKey())
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

func (m *Map) Erase() {
	// TODO needs DelPrefix method in KVStore
	panic("implement me")
}

// Iterate non-deterministic
func (m *ImmutableMap) Iterate(f func(elemKey []byte, value []byte) bool) error {
	prefix := m.getElemKey(nil)
	return m.kvr.Iterate(prefix, func(key kv.Key, value []byte) bool {
		return f([]byte(key)[len(prefix):], value)
		//return f([]byte(key), value)
	})
}

// Iterate non-deterministic
func (m *ImmutableMap) IterateKeys(f func(elemKey []byte) bool) error {
	prefix := m.getElemKey(nil)
	return m.kvr.IterateKeys(prefix, func(key kv.Key) bool {
		return f([]byte(key)[len(prefix):])
	})
}

// Iterate non-deterministic
func (m *ImmutableMap) MustIterate(f func(elemKey []byte, value []byte) bool) {
	err := m.Iterate(f)
	if err != nil {
		panic(err)
	}
}

// Iterate non-deterministic
func (m *ImmutableMap) MustIterateKeys(f func(elemKey []byte) bool) {
	err := m.IterateKeys(f)
	if err != nil {
		panic(err)
	}
}

func (m *ImmutableMap) IterateBalances(f func(color ledgerstate.Color, bal uint64) bool) error {
	var err error
	m.MustIterate(func(elemKey []byte, value []byte) bool {
		col, _, err := ledgerstate.ColorFromBytes(elemKey)
		if err != nil {
			return false
		}
		v, err := util.Uint64From8Bytes(value)
		if err != nil {
			return false
		}
		return f(col, v)
	})
	return err
}
