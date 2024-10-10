package orderedmap

import (
	"github.com/mitchellh/hashstructure/v2"
)

type OrderedMap[K comparable, V any] struct {
	m           map[uint64]V
	insertOrder []K
}

func New[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		m:           make(map[uint64]V),
		insertOrder: []K{},
	}
}

// Set method to add or update a key-value pair and maintain insertion order
func (m *OrderedMap[K, V]) Set(key K, value V) {
	hash := GetHash(key)
	if _, exists := m.m[hash]; !exists {
		m.insertOrder = append(m.insertOrder, key)
	}
	m.m[hash] = value
}

// Insert method to add or update a key-value pair and maintain insertion order
func (m *OrderedMap[K, V]) Insert(key K, value V) {
	m.InsertFull(key, value)
}

// Insert method to add or update a key-value pair and maintain insertion order and
func (m *OrderedMap[K, V]) InsertFull(key K, value V) int {
	hash := GetHash(key)
	_, exists := m.m[hash]
	m.m[hash] = value
	if !exists {
		m.insertOrder = append(m.insertOrder, key)
		return len(m.insertOrder) - 1
	}

	idx, exists := m.Find(key)
	if !exists {
		panic("key didn't successfully add")
	}
	return idx
}

// Get method to retrieve the value for a given key
func (m *OrderedMap[K, V]) Get(key K) (V, bool) {
	hash := GetHash(key)
	value, exists := m.m[hash]
	return value, exists
}

func (m *OrderedMap[K, V]) Find(key K) (int, bool) {
	for i, v := range m.insertOrder {
		if v == key {
			return i, true
		}
	}
	return 0, false
}

func (m *OrderedMap[K, V]) Index(i int) (*K, bool) {
	if len(m.insertOrder)-1 < i {
		k := m.insertOrder[i]
		return &k, true
	}
	return nil, false
}

func (m *OrderedMap[K, V]) Len() int {
	return len(m.m)
}

// ForEach method to iterate over all key-value pairs in insertion order
func (m *OrderedMap[K, V]) ForEach(action func(K, V)) {
	for _, key := range m.insertOrder {
		hash := GetHash(key)
		action(key, m.m[hash])
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
