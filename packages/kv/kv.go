package kv

import (
	"fmt"
)

// Key represents a key in the KVStore, to avoid unnecessary conversions
// between string and []byte, we use string as key data type, but it does
// not necessarily have to be a valid UTF-8 string.
type Key string

func (k Key) Hex() string {
	return fmt.Sprintf("kv.Key('%X')", k)
}

const EmptyPrefix = Key("")

func (k Key) HasPrefix(prefix Key) bool {
	if len(prefix) > len(k) {
		return false
	}
	return k[:len(prefix)] == prefix
}

// KVStore represents a key-value store
// where both keys and values are arbitrary byte slices.
type KVStore interface {
	KVWriter
	KVStoreReader
}

type KVReader interface {
	// Get returns the value, or nil if not found
	Get(key Key) []byte
	Has(key Key) bool
}

type KVWriter interface {
	Set(key Key, value []byte)
	Del(key Key)
}

type KVIterator interface {
	Iterate(prefix Key, f func(key Key, value []byte) bool)
	IterateKeys(prefix Key, f func(key Key) bool)
	IterateSorted(prefix Key, f func(key Key, value []byte) bool)
	IterateKeysSorted(prefix Key, f func(key Key) bool)
}

type KVStoreReader interface {
	KVReader
	KVIterator
}
