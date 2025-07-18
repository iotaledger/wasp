// Package subrealm provides functionality for working with isolated subsets
// of the key-value store. It enables creation of distinct realms within a
// single KV store, allowing for better organization and isolation of data.
package subrealm

import (
	"github.com/iotaledger/wasp/v2/packages/kv"
)

type subrealm struct {
	kv     kv.KVStore
	prefix kv.Key
}

func New(kvStore kv.KVStore, prefix kv.Key) kv.KVStore {
	return &subrealm{kvStore, prefix}
}

func (s *subrealm) Set(key kv.Key, value []byte) {
	s.kv.Set(s.prefix+key, value)
}

func (s *subrealm) Del(key kv.Key) {
	s.kv.Del(s.prefix + key)
}

// Get returns the value, or nil if not found
func (s *subrealm) Get(key kv.Key) []byte {
	return s.kv.Get(s.prefix + key)
}

func (s *subrealm) Has(key kv.Key) bool {
	return s.kv.Has(s.prefix + key)
}

func (s *subrealm) Iterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	s.kv.Iterate(s.prefix+prefix, func(key kv.Key, value []byte) bool {
		return f(key[len(s.prefix):], value)
	})
}

func (s *subrealm) IterateKeys(prefix kv.Key, f func(key kv.Key) bool) {
	s.kv.IterateKeys(s.prefix+prefix, func(key kv.Key) bool {
		return f(key[len(s.prefix):])
	})
}

func (s *subrealm) IterateSorted(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	s.kv.IterateSorted(s.prefix+prefix, func(key kv.Key, value []byte) bool {
		return f(key[len(s.prefix):], value)
	})
}

func (s *subrealm) IterateKeysSorted(prefix kv.Key, f func(key kv.Key) bool) {
	s.kv.IterateKeysSorted(s.prefix+prefix, func(key kv.Key) bool {
		return f(key[len(s.prefix):])
	})
}
