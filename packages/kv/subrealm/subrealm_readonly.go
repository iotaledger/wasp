package subrealm

import (
	"github.com/iotaledger/wasp/v2/packages/kv"
)

type subrealmReadOnly struct {
	kv     kv.KVStoreReader
	prefix kv.Key
}

func NewReadOnly(kvReader kv.KVStoreReader, prefix kv.Key) kv.KVStoreReader {
	return &subrealmReadOnly{kvReader, prefix}
}

// Get returns the value, or nil if not found
func (s *subrealmReadOnly) Get(key kv.Key) []byte {
	return s.kv.Get(s.prefix + key)
}

func (s *subrealmReadOnly) Has(key kv.Key) bool {
	return s.kv.Has(s.prefix + key)
}

func (s *subrealmReadOnly) Iterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	s.kv.Iterate(s.prefix+prefix, func(key kv.Key, value []byte) bool {
		return f(key[len(s.prefix):], value)
	})
}

func (s *subrealmReadOnly) IterateKeys(prefix kv.Key, f func(key kv.Key) bool) {
	s.kv.IterateKeys(s.prefix+prefix, func(key kv.Key) bool {
		return f(key[len(s.prefix):])
	})
}

func (s *subrealmReadOnly) IterateSorted(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	s.kv.IterateSorted(s.prefix+prefix, func(key kv.Key, value []byte) bool {
		return f(key[len(s.prefix):], value)
	})
}

func (s *subrealmReadOnly) IterateKeysSorted(prefix kv.Key, f func(key kv.Key) bool) {
	s.kv.IterateKeysSorted(s.prefix+prefix, func(key kv.Key) bool {
		return f(key[len(s.prefix):])
	})
}
