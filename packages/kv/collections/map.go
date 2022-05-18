package collections

import (
	"errors"

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

func (m *Map) addToSize(amount int) error {
	n, err := m.Len()
	if err != nil {
		return err
	}
	n = uint32(int(n) + amount)
	if n == 0 {
		m.kvw.Del(kv.Key(MapSizeKey(m.name)))
	} else {
		m.kvw.Set(kv.Key(MapSizeKey(m.name)), util.Uint32To4Bytes(n))
	}
	return nil
}

func (m *ImmutableMap) GetAt(key []byte) ([]byte, error) {
	ret, err := m.kvr.Get(kv.Key(MapElemKey(m.name, key)))
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

func (m *Map) SetAt(key, value []byte) error {
	keyExists, err := m.HasAt(key)
	if err != nil {
		return err
	}
	if !keyExists {
		err = m.addToSize(1)
		if err != nil {
			return err
		}
	}
	m.kvw.Set(kv.Key(MapElemKey(m.name, key)), value)
	return nil
}

func (m *Map) MustSetAt(key, value []byte) {
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
	m.kvw.Del(kv.Key(MapElemKey(m.name, key)))
	return nil
}

func (m *Map) MustDelAt(key []byte) {
	err := m.DelAt(key)
	if err != nil {
		panic(err)
	}
}

func (m *ImmutableMap) HasAt(key []byte) (bool, error) {
	return m.kvr.Has(kv.Key(MapElemKey(m.name, key)))
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
	v, err := m.kvr.Get(kv.Key(MapSizeKey(m.name)))
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

// Erase the map.
func (m *Map) Erase() {
	m.MustIterateKeys(func(elemKey []byte) bool {
		m.MustDelAt(elemKey)
		return true
	})
}

// Iterate non-deterministic
func (m *ImmutableMap) Iterate(f func(elemKey []byte, value []byte) bool) error {
	prefix := kv.Key(MapElemKey(m.name, nil))
	return m.kvr.Iterate(prefix, func(key kv.Key, value []byte) bool {
		return f([]byte(key)[len(prefix):], value)
	})
}

// IterateKeys non-deterministic
func (m *ImmutableMap) IterateKeys(f func(elemKey []byte) bool) error {
	prefix := kv.Key(MapElemKey(m.name, nil))
	return m.kvr.IterateKeys(prefix, func(key kv.Key) bool {
		return f([]byte(key)[len(prefix):])
	})
}

// MustIterate non-deterministic
func (m *ImmutableMap) MustIterate(f func(elemKey []byte, value []byte) bool) {
	err := m.Iterate(f)
	if err != nil {
		panic(err)
	}
}

// MustIterateKeys non-deterministic
func (m *ImmutableMap) MustIterateKeys(f func(elemKey []byte) bool) {
	err := m.IterateKeys(f)
	if err != nil {
		panic(err)
	}
}
