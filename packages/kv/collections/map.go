package collections

import (
	"math"

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

// For easy distinction between arrays and map collections
// we use '#' as separator for arrays and '.' for maps.
// Do not change this value unless you want to break how
// WasmLib maps these collections in the exact same way
const (
	mapElemKeyCode = byte('.')
	mapSizeKeyCode = byte('#')
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

func MapSizeKey(mapName string) []byte {
	ret := make([]byte, 0, len(mapName)+1)
	ret = append(ret, []byte(mapName)...)
	return append(ret, mapSizeKeyCode)
}

func MapElemKey(mapName string, keyInMap []byte) []byte {
	ret := make([]byte, 0, len(mapName)+len(keyInMap)+1)
	ret = append(ret, []byte(mapName)...)
	ret = append(ret, mapElemKeyCode)
	return append(ret, keyInMap...)
}

func (m *Map) Immutable() *ImmutableMap {
	return m.ImmutableMap
}

func (m *ImmutableMap) Name() string {
	return m.name
}

func (m *Map) addToSize(amount int) {
	n := int64(m.Len()) + int64(amount)
	if n < 0 {
		panic("negative size in Map")
	}
	if n > math.MaxUint32 {
		panic("Map is full")
	}
	if n == 0 {
		m.kvw.Del(kv.Key(MapSizeKey(m.name)))
	} else {
		m.kvw.Set(kv.Key(MapSizeKey(m.name)), util.Uint32To4Bytes(uint32(n)))
	}
}

func (m *ImmutableMap) GetAt(key []byte) []byte {
	return m.kvr.Get(kv.Key(MapElemKey(m.name, key)))
}

func (m *Map) SetAt(key, value []byte) {
	if !m.HasAt(key) {
		m.addToSize(1)
	}
	m.kvw.Set(kv.Key(MapElemKey(m.name, key)), value)
}

func (m *Map) DelAt(key []byte) {
	if !m.HasAt(key) {
		return
	}
	m.addToSize(-1)
	m.kvw.Del(kv.Key(MapElemKey(m.name, key)))
}

func (m *ImmutableMap) HasAt(key []byte) bool {
	return m.kvr.Has(kv.Key(MapElemKey(m.name, key)))
}

func (m *ImmutableMap) Len() uint32 {
	v := m.kvr.Get(kv.Key(MapSizeKey(m.name)))
	if v == nil {
		return 0
	}
	return util.MustUint32From4Bytes(v)
}

// Erase the map.
func (m *Map) Erase() {
	m.IterateKeys(func(elemKey []byte) bool {
		m.DelAt(elemKey)
		return true
	})
}

// Iterate non-deterministic
func (m *ImmutableMap) Iterate(f func(elemKey []byte, value []byte) bool) {
	prefix := kv.Key(MapElemKey(m.name, nil))
	m.kvr.Iterate(prefix, func(key kv.Key, value []byte) bool {
		return f([]byte(key)[len(prefix):], value)
	})
}

// IterateKeys non-deterministic
func (m *ImmutableMap) IterateKeys(f func(elemKey []byte) bool) {
	prefix := kv.Key(MapElemKey(m.name, nil))
	m.kvr.IterateKeys(prefix, func(key kv.Key) bool {
		return f([]byte(key)[len(prefix):])
	})
}
