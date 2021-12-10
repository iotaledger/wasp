package subrealm

import (
	"github.com/iotaledger/wasp/packages/iscp/gas"
	"github.com/iotaledger/wasp/packages/kv"
)

type BurnGasFn = func(gas uint64)

type subrealm struct {
	kv      kv.KVStore
	prefix  kv.Key
	burnGas BurnGasFn
}

func New(burnGas BurnGasFn, kvStore kv.KVStore, prefix kv.Key) kv.KVStore {
	return &subrealm{kvStore, prefix, burnGas}
}

func (s *subrealm) Set(key kv.Key, value []byte) {
	s.burnGas(gas.StoreBytes(len(key) + len(value)))
	s.kv.Set(s.prefix+key, value)
}

func (s *subrealm) Del(key kv.Key) {
	// TODO should Deletions from the state burn gas?
	s.kv.Del(s.prefix + key)
}

// Get returns the value, or nil if not found
func (s *subrealm) Get(key kv.Key) ([]byte, error) {
	s.burnGas(gas.ReadTheState(len(key)))
	return s.kv.Get(s.prefix + key)
}

func (s *subrealm) Has(key kv.Key) (bool, error) {
	s.burnGas(gas.ReadTheState(len(key)))
	return s.kv.Has(s.prefix + key)
}

func (s *subrealm) Iterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) error {
	return s.kv.Iterate(s.prefix+prefix, func(key kv.Key, value []byte) bool {
		s.burnGas(gas.ReadTheState(len(key) + len(value)))
		return f(key[len(s.prefix):], value)
	})
}

func (s *subrealm) IterateKeys(prefix kv.Key, f func(key kv.Key) bool) error {
	return s.kv.IterateKeys(s.prefix+prefix, func(key kv.Key) bool {
		s.burnGas(gas.ReadTheState(len(key)))
		return f(key[len(s.prefix):])
	})
}

func (s *subrealm) IterateSorted(prefix kv.Key, f func(key kv.Key, value []byte) bool) error {
	return s.kv.IterateSorted(s.prefix+prefix, func(key kv.Key, value []byte) bool {
		s.burnGas(gas.ReadTheState(len(key) + len(value)))
		return f(key[len(s.prefix):], value)
	})
}

func (s *subrealm) IterateKeysSorted(prefix kv.Key, f func(key kv.Key) bool) error {
	return s.kv.IterateKeysSorted(s.prefix+prefix, func(key kv.Key) bool {
		s.burnGas(gas.ReadTheState(len(key)))
		return f(key[len(s.prefix):])
	})
}

func (s *subrealm) MustGet(key kv.Key) []byte {
	s.burnGas(gas.ReadTheState(len(key)))
	return kv.MustGet(s, key)
}

func (s *subrealm) MustHas(key kv.Key) bool {
	s.burnGas(gas.ReadTheState(len(key)))
	return kv.MustHas(s, key)
}

func (s *subrealm) MustIterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	kv.MustIterate(s, prefix, s.iterateKVFnWithGasBurn(f))
}

func (s *subrealm) MustIterateKeys(prefix kv.Key, f func(key kv.Key) bool) {
	kv.MustIterateKeys(s, prefix, s.iterateKeysFnWithGasBurn(f))
}

func (s *subrealm) MustIterateSorted(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	kv.MustIterateSorted(s, prefix, s.iterateKVFnWithGasBurn(f))
}

func (s *subrealm) MustIterateKeysSorted(prefix kv.Key, f func(key kv.Key) bool) {
	kv.MustIterateKeysSorted(s, prefix, s.iterateKeysFnWithGasBurn(f))
}

func (s *subrealm) iterateKeysFnWithGasBurn(f func(key kv.Key) bool) func(key kv.Key) bool {
	return func(key kv.Key) bool {
		s.burnGas(gas.ReadTheState(len(key)))
		return f(key)
	}
}

func (s *subrealm) iterateKVFnWithGasBurn(f func(key kv.Key, value []byte) bool) func(key kv.Key, value []byte) bool {
	return func(key kv.Key, value []byte) bool {
		s.burnGas(gas.ReadTheState(len(key) + len(value)))
		return f(key, value)
	}
}
