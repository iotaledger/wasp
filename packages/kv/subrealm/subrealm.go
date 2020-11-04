package subrealm

import "github.com/iotaledger/wasp/packages/kv"

type subrealm struct {
	kv     kv.KVStore
	prefix kv.Key
}

func New(kv kv.KVStore, prefix kv.Key) kv.KVStore {
	return &subrealm{kv, prefix}
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
