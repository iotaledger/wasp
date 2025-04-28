// Package collections provides high-level data structures built on top of
// the kv package. It implements common collection types such as maps, arrays,
// and other data structures with a key-value store backend.
package collections

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util/rwutil"
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

// For easy distinction between arrays and map collections
// we use '#' as separator for arrays and '.' for maps.
const (
	mapElemKeyCode = byte('.')
)

func NewMap(kvStore kv.KVStore, name string) *Map {
	return &Map{
		ImmutableMap: NewMapReadOnly(kvStore, name),
		kvw:          kvStore,
	}
}

func NewMapReadOnly(kvReader kv.KVStoreReader, name string) *ImmutableMap {
	return &ImmutableMap{
		kvr:  kvReader,
		name: name,
	}
}

func MapElemKey(name string, key []byte) kv.Key {
	return kv.Key(rwutil.NewBytesWriter().
		WriteN([]byte(name)).
		WriteByte(mapElemKeyCode).
		WriteN(key).
		Bytes())
}

func (m *Map) Immutable() *ImmutableMap {
	return m.ImmutableMap
}

func (m *ImmutableMap) Name() string {
	return m.name
}

func (m *Map) addToSize(amount int) {
	key := kv.Key(m.name)
	data := m.kvr.Get(key)
	if data != nil {
		amount += rwutil.NewBytesReader(data).Must().ReadSize32()
	}
	if amount == 0 {
		m.kvw.Del(key)
		return
	}
	m.kvw.Set(key, rwutil.NewBytesWriter().WriteSize32(amount).Bytes())
}

func (m *ImmutableMap) GetAt(key []byte) []byte {
	return m.kvr.Get(MapElemKey(m.name, key))
}

func (m *Map) SetAt(key, value []byte) {
	if !m.HasAt(key) {
		m.addToSize(1)
	}
	m.kvw.Set(MapElemKey(m.name, key), value)
}

func (m *Map) DelAt(key []byte) {
	if !m.HasAt(key) {
		return
	}
	m.addToSize(-1)
	m.kvw.Del(MapElemKey(m.name, key))
}

func (m *ImmutableMap) HasAt(key []byte) bool {
	return m.kvr.Has(MapElemKey(m.name, key))
}

func (m *ImmutableMap) Len() uint32 {
	data := m.kvr.Get(kv.Key(m.name))
	if data == nil {
		return 0
	}
	return uint32(rwutil.NewBytesReader(data).Must().ReadSize32())
}

func (m *Map) Keys() [][]byte {
	var keys [][]byte
	m.IterateKeys(func(elemKey []byte) bool {
		keys = append(keys, elemKey)
		return true
	})
	return keys
}

func (m *ImmutableMap) Keys() [][]byte {
	var keys [][]byte
	m.IterateKeys(func(elemKey []byte) bool {
		keys = append(keys, elemKey)
		return true
	})
	return keys
}

// Erase the map.
func (m *Map) Erase() {
	for _, k := range m.Keys() {
		m.DelAt(k)
	}
}

// Iterate non-deterministic
func (m *ImmutableMap) Iterate(f func(elemKey []byte, value []byte) bool) {
	prefix := MapElemKey(m.name, nil)
	m.kvr.Iterate(prefix, func(key kv.Key, value []byte) bool {
		return f([]byte(key)[len(prefix):], value)
	})
}

// IterateKeys non-deterministic
func (m *ImmutableMap) IterateKeys(f func(elemKey []byte) bool) {
	prefix := MapElemKey(m.name, nil)
	m.kvr.IterateKeys(prefix, func(key kv.Key) bool {
		return f([]byte(key)[len(prefix):])
	})
}
