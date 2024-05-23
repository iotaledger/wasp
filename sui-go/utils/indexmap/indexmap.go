package indexmap

import "github.com/mitchellh/hashstructure/v2"

type IndexMap[K comparable, V any] struct {
	UnindexedMap    map[K]V
	InsertOrderList []K
}

func NewIndexMap[K comparable, V any]() *IndexMap[K, V] {
	return &IndexMap[K, V]{
		UnindexedMap:    make(map[K]V),
		InsertOrderList: []K{},
	}
}

// Set method to add or update a key-value pair and maintain insertion order
func (m *IndexMap[K, V]) Set(key K, value V) {
	if _, exists := m.UnindexedMap[key]; !exists {
		m.InsertOrderList = append(m.InsertOrderList, key)
	}
	m.UnindexedMap[key] = value
}

// Insert method to add or update a key-value pair and maintain insertion order
func (m *IndexMap[K, V]) Insert(key K, value V) {
	m.Set(key, value)
}

// Insert method to add or update a key-value pair and maintain insertion order and
func (m *IndexMap[K, V]) InsertFull(key K, value V) int {
	if _, exists := m.UnindexedMap[key]; !exists {
		m.InsertOrderList = append(m.InsertOrderList, key)
		return len(m.InsertOrderList) - 1
	}
	m.UnindexedMap[key] = value
	idx, exists := m.Index(key)
	if !exists {
		panic("key didn't successfully add")
	}
	return idx
}

// Get method to retrieve the value for a given key
func (m *IndexMap[K, V]) Get(key K) (V, bool) {
	value, exists := m.UnindexedMap[key]
	return value, exists
}

func (m *IndexMap[K, V]) Index(key K) (int, bool) {
	for i, v := range m.InsertOrderList {
		if v == key {
			return i, true
		}
	}
	return 0, false
}

func (m *IndexMap[K, V]) Len() int {
	return len(m.UnindexedMap)
}

// ForEach method to iterate over all key-value pairs in insertion order
func (m *IndexMap[K, V]) ForEach(action func(K, V)) {
	for _, key := range m.InsertOrderList {
		action(key, m.UnindexedMap[key])
	}
}

// return a deterministic hash of struct/any datatype
func GetHash(obj any) uint64 {
	// TODO we can implement our own hash func for go structs
	hash, err := hashstructure.Hash(obj, hashstructure.FormatV2, nil)
	if err != nil {
		panic(err)
	}
	return hash
}
