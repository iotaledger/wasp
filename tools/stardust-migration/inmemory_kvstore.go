package main

import (
	"bytes"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
)

func NewInMemoryKVStore(onlyEffectiveMuts, deleteIfNotSet, readOnlyUncommitted bool) *InMemoryKVStore {
	committed := buffered.NewBufferedKVStore(NoopKVStoreReader[kv.Key]{})
	uncommitted := buffered.NewBufferedKVStore(committed)

	if readOnlyUncommitted {
		uncommitted = buffered.NewBufferedKVStore(NoopKVStoreReader[kv.Key]{})
	}

	return &InMemoryKVStore{
		committed:         committed,
		uncommitted:       uncommitted,
		onlyEffectiveMuts: onlyEffectiveMuts,
		deleteIfNotSet:    deleteIfNotSet,
	}
}

type InMemoryKVStore struct {
	committed         *buffered.BufferedKVStore
	uncommitted       *buffered.BufferedKVStore
	onlyEffectiveMuts bool
	deleteIfNotSet    bool
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
	if b.deleteIfNotSet {
		committedSets := b.committed.Mutations().Sets
		uncommittedSets := b.uncommitted.Mutations().Sets
		for key := range committedSets {
			_, ok := uncommittedSets[key]
			if !ok {
				b.uncommitted.Del(key)
			}
		}
	}

	if b.onlyEffectiveMuts {
		committedSets := b.committed.Mutations().Sets
		uncommittedSets := b.uncommitted.Mutations().Sets

		for key, value := range uncommittedSets {
			committed, ok := committedSets[key]
			if ok && bytes.Equal(committed, value) {
				delete(uncommittedSets, key)
			}
		}

		uncommittedDels := b.uncommitted.Mutations().Dels

		for key := range uncommittedDels {
			_, ok := committedSets[key]
			if !ok {
				delete(uncommittedDels, key)
			}
		}
	}

	muts := b.uncommitted.Mutations()
	muts.ApplyTo(b.committed)
	b.uncommitted.SetMutations(buffered.NewMutations())

	return muts
}

func (b *InMemoryKVStore) Set(key kv.Key, value []byte) {
	b.uncommitted.Set(key, value)
}

func (b *InMemoryKVStore) Del(key kv.Key) {
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
