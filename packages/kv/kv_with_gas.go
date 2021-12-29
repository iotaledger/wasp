package kv

import (
	"github.com/iotaledger/wasp/packages/vm/gas"
)

type BurnGasFn = func(gas uint64)

type kvWithGas struct {
	kv      KVStore
	burnGas BurnGasFn
}

func WithGas(kv KVStore, burnGas BurnGasFn) KVStore {
	return &kvWithGas{kv, burnGas}
}

func (s *kvWithGas) Set(key Key, value []byte) {
	s.burnGas(gas.StoreBytes(len(key) + len(value)))
	s.kv.Set(key, value)
}

func (s *kvWithGas) Del(key Key) {
	// TODO should Deletions from the state burn gas?
	s.kv.Del(key)
}

// Get returns the value, or nil if not found
func (s *kvWithGas) Get(key Key) ([]byte, error) {
	s.burnGas(gas.ReadTheState(len(key)))
	return s.kv.Get(key)
}

func (s *kvWithGas) Has(key Key) (bool, error) {
	s.burnGas(gas.ReadTheState(len(key)))
	return s.kv.Has(key)
}

func (s *kvWithGas) Iterate(prefix Key, f func(key Key, value []byte) bool) error {
	return s.kv.Iterate(prefix, s.iterateKVFnWithGasBurn(f))
}

func (s *kvWithGas) IterateKeys(prefix Key, f func(key Key) bool) error {
	return s.kv.IterateKeys(prefix, s.iterateKeysFnWithGasBurn(f))
}

func (s *kvWithGas) IterateSorted(prefix Key, f func(key Key, value []byte) bool) error {
	return s.kv.IterateSorted(prefix, s.iterateKVFnWithGasBurn(f))
}

func (s *kvWithGas) IterateKeysSorted(prefix Key, f func(key Key) bool) error {
	return s.kv.IterateKeysSorted(prefix, s.iterateKeysFnWithGasBurn(f))
}

func (s *kvWithGas) MustGet(key Key) []byte {
	s.burnGas(gas.ReadTheState(len(key)))
	return s.kv.MustGet(key)
}

func (s *kvWithGas) MustHas(key Key) bool {
	s.burnGas(gas.ReadTheState(len(key)))
	return s.kv.MustHas(key)
}

func (s *kvWithGas) MustIterate(prefix Key, f func(key Key, value []byte) bool) {
	s.kv.MustIterate(prefix, s.iterateKVFnWithGasBurn(f))
}

func (s *kvWithGas) MustIterateKeys(prefix Key, f func(key Key) bool) {
	s.kv.MustIterateKeys(prefix, s.iterateKeysFnWithGasBurn(f))
}

func (s *kvWithGas) MustIterateSorted(prefix Key, f func(key Key, value []byte) bool) {
	s.kv.MustIterateSorted(prefix, s.iterateKVFnWithGasBurn(f))
}

func (s *kvWithGas) MustIterateKeysSorted(prefix Key, f func(key Key) bool) {
	s.kv.MustIterateKeysSorted(prefix, s.iterateKeysFnWithGasBurn(f))
}

func (s *kvWithGas) iterateKeysFnWithGasBurn(f func(key Key) bool) func(key Key) bool {
	return func(key Key) bool {
		s.burnGas(gas.ReadTheState(len(key)))
		return f(key)
	}
}

func (s *kvWithGas) iterateKVFnWithGasBurn(f func(key Key, value []byte) bool) func(key Key, value []byte) bool {
	return func(key Key, value []byte) bool {
		s.burnGas(gas.ReadTheState(len(key) + len(value)))
		return f(key, value)
	}
}
