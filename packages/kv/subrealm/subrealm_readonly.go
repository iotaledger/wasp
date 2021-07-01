package subrealm

import (
	"github.com/iotaledger/wasp/packages/kv"
)

type subrealmReadOnly struct {
	kv     kv.KVStoreReader
	prefix kv.Key
}

func NewReadOnly(kvReader kv.KVStoreReader, prefix kv.Key) kv.KVStoreReader {
	return &subrealmReadOnly{kvReader, prefix}
}

// Get returns the value, or nil if not found
func (s *subrealmReadOnly) Get(key kv.Key) ([]byte, error) {
	return s.kv.Get(s.prefix + key)
}

func (s *subrealmReadOnly) Has(key kv.Key) (bool, error) {
	return s.kv.Has(s.prefix + key)
}

func (s *subrealmReadOnly) Iterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) error {
	return s.kv.Iterate(s.prefix+prefix, func(key kv.Key, value []byte) bool {
		return f(key[len(s.prefix):], value)
	})
}

func (s *subrealmReadOnly) IterateKeys(prefix kv.Key, f func(key kv.Key) bool) error {
	return s.kv.IterateKeys(s.prefix+prefix, func(key kv.Key) bool {
		return f(key[len(s.prefix):])
	})
}

func (s *subrealmReadOnly) IterateSorted(prefix kv.Key, f func(key kv.Key, value []byte) bool) error {
	return s.kv.IterateSorted(s.prefix+prefix, func(key kv.Key, value []byte) bool {
		return f(key[len(s.prefix):], value)
	})
}

func (s *subrealmReadOnly) IterateKeysSorted(prefix kv.Key, f func(key kv.Key) bool) error {
	return s.kv.IterateKeysSorted(s.prefix+prefix, func(key kv.Key) bool {
		return f(key[len(s.prefix):])
	})
}

func (s *subrealmReadOnly) MustGet(key kv.Key) []byte {
	return kv.MustGet(s, key)
}

func (s *subrealmReadOnly) MustHas(key kv.Key) bool {
	return kv.MustHas(s, key)
}

func (s *subrealmReadOnly) MustIterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	kv.MustIterate(s, prefix, f)
}

func (s *subrealmReadOnly) MustIterateKeys(prefix kv.Key, f func(key kv.Key) bool) {
	kv.MustIterateKeys(s, prefix, f)
}

func (s *subrealmReadOnly) MustIterateSorted(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	kv.MustIterateSorted(s, prefix, f)
}

func (s *subrealmReadOnly) MustIterateKeysSorted(prefix kv.Key, f func(key kv.Key) bool) {
	kv.MustIterateKeysSorted(s, prefix, f)
}
