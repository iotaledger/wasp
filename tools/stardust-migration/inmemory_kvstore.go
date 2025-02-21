package main

import (
	"bytes"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
)

func NewInMemoryKVStore(onlyEffectiveMutations bool) *InMemoryKVStore {
	committed := buffered.NewBufferedKVStore(NoopKVStoreReader[kv.Key]{})
	uncommitted := buffered.NewBufferedKVStore(committed)

	return &InMemoryKVStore{
		committed:              committed,
		uncommitted:            uncommitted,
		onlyEffectiveMutations: onlyEffectiveMutations,
	}
}

type InMemoryKVStore struct {
	committed              *buffered.BufferedKVStore
	uncommitted            *buffered.BufferedKVStore
	onlyEffectiveMutations bool
}

var _ kv.KVStoreReader = &InMemoryKVStore{}
var _ kv.KVStore = &InMemoryKVStore{}

func (b *InMemoryKVStore) CommittedSize() int {
	return len(b.committed.Mutations().Sets)
}

func (b *InMemoryKVStore) Mutations() *buffered.Mutations {
	return b.uncommitted.Mutations()
}

func (b *InMemoryKVStore) MutationsCount() int {
	return len(b.uncommitted.Mutations().Sets) + len(b.uncommitted.Mutations().Dels)
}

func (b *InMemoryKVStore) Commit() *buffered.Mutations {
	muts := b.uncommitted.Mutations()
	muts.ApplyTo(b.committed)
	b.uncommitted = buffered.NewBufferedKVStore(b.committed)
	return muts
}

func (b *InMemoryKVStore) Set(key kv.Key, value []byte) {
	if b.onlyEffectiveMutations {
		committed := b.committed.Get(key)
		if bytes.Equal(committed, value) {
			delete(b.uncommitted.Mutations().Dels, key)
			return
		}
	}

	b.uncommitted.Set(key, value)
}

func (b *InMemoryKVStore) Del(key kv.Key) {
	if b.onlyEffectiveMutations {
		if !b.committed.Has(key) {
			delete(b.uncommitted.Mutations().Sets, key)
			return
		}
	}

	b.uncommitted.Del(key)
}

func (b *InMemoryKVStore) Get(key kv.Key) []byte {
	return b.uncommitted.Get(key)
}

func (b *InMemoryKVStore) Has(key kv.Key) bool {
	return b.uncommitted.Has(key)
}

func (b *InMemoryKVStore) Iterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	b.uncommitted.Iterate(prefix, f)
}

func (b *InMemoryKVStore) IterateKeys(prefix kv.Key, f func(key kv.Key) bool) {
	b.uncommitted.IterateKeys(prefix, f)
}

func (b *InMemoryKVStore) IterateSorted(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	b.uncommitted.IterateSorted(prefix, f)
}

func (b *InMemoryKVStore) IterateKeysSorted(prefix kv.Key, f func(key kv.Key) bool) {
	b.uncommitted.IterateKeysSorted(prefix, f)
}
