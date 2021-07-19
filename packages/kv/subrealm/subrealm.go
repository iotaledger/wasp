package subrealm

import (
	"github.com/iotaledger/wasp/packages/kv"
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
func (s *subrealm) Get(key kv.Key) ([]byte, error) {
	return s.kv.Get(s.prefix + key)
}

func (s *subrealm) Has(key kv.Key) (bool, error) {
	return s.kv.Has(s.prefix + key)
}

func (s *subrealm) Iterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) error {
	return s.kv.Iterate(s.prefix+prefix, func(key kv.Key, value []byte) bool {
		return f(key[len(s.prefix):], value)
	})
}

func (s *subrealm) IterateKeys(prefix kv.Key, f func(key kv.Key) bool) error {
	return s.kv.IterateKeys(s.prefix+prefix, func(key kv.Key) bool {
		return f(key[len(s.prefix):])
	})
}

func (s *subrealm) IterateSorted(prefix kv.Key, f func(key kv.Key, value []byte) bool) error {
	return s.kv.IterateSorted(s.prefix+prefix, func(key kv.Key, value []byte) bool {
		return f(key[len(s.prefix):], value)
	})
}

func (s *subrealm) IterateKeysSorted(prefix kv.Key, f func(key kv.Key) bool) error {
	return s.kv.IterateKeysSorted(s.prefix+prefix, func(key kv.Key) bool {
		return f(key[len(s.prefix):])
	})
}

func (s *subrealm) MustGet(key kv.Key) []byte {
	return kv.MustGet(s, key)
}

func (s *subrealm) MustHas(key kv.Key) bool {
	return kv.MustHas(s, key)
}

func (s *subrealm) MustIterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	kv.MustIterate(s, prefix, f)
}

func (s *subrealm) MustIterateKeys(prefix kv.Key, f func(key kv.Key) bool) {
	kv.MustIterateKeys(s, prefix, f)
}

func (s *subrealm) MustIterateSorted(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	kv.MustIterateSorted(s, prefix, f)
}

func (s *subrealm) MustIterateKeysSorted(prefix kv.Key, f func(key kv.Key) bool) {
	kv.MustIterateKeysSorted(s, prefix, f)
}
