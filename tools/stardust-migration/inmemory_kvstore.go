package main

import (
	"bytes"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
)

func NewInMemoryKVStore(readOnlyUncommitted bool) *InMemoryKVStore {
	committed := buffered.NewBufferedKVStore(NoopKVStoreReader[kv.Key]{})
	uncommitted := buffered.NewBufferedKVStore(committed)

	if readOnlyUncommitted {
		uncommitted = buffered.NewBufferedKVStore(NoopKVStoreReader[kv.Key]{})
	}

	return &InMemoryKVStore{
		committed:        committed,
		uncommitted:      uncommitted,
		committedMarked:  make(map[kv.Key]struct{}),
		uncommitedMarked: make(map[kv.Key]struct{}),
	}
}

type InMemoryKVStore struct {
	committed        *buffered.BufferedKVStore
	uncommitted      *buffered.BufferedKVStore
	marking          bool
	uncommitedMarked map[kv.Key]struct{}
	committedMarked  map[kv.Key]struct{}
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

func (b *InMemoryKVStore) StartMarking() {
	b.marking = true
}

func (b *InMemoryKVStore) StopMarking() {
	b.marking = false
}

// DeleteIfNotSet deletes marked entries if they where not set since the last commit
func (b *InMemoryKVStore) DeleteMarkedIfNotSet() {
	uncommittedSets := b.uncommitted.Mutations().Sets

	for key := range b.committedMarked {
		if _, isSet := uncommittedSets[key]; !isSet {
			b.uncommitted.Del(key)
		}
	}
}

// DeleteIfNotSet deletes entries if they where not set since the last commit
func (b *InMemoryKVStore) DeleteIfNotSet() {
	uncommittedSets := b.uncommitted.Mutations().Sets

	for key := range b.committed.Mutations().Sets {
		if _, isSet := uncommittedSets[key]; !isSet {
			b.uncommitted.Del(key)
		}
	}
}

// RemoveRedundantMutations removes mutations that have no effect
func (b *InMemoryKVStore) RemoveRedundantMutations() {
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

func (b *InMemoryKVStore) Commit(onlyEffectiveMutations bool) *buffered.Mutations {
	if onlyEffectiveMutations {
		b.RemoveRedundantMutations()
	}

	muts := b.uncommitted.Mutations()
	muts.ApplyTo(b.committed)
	b.uncommitted.SetMutations(buffered.NewMutations())
	b.committedMarked = b.uncommitedMarked
	b.uncommitedMarked = make(map[kv.Key]struct{}, len(b.uncommitedMarked))

	return muts
}

func (b *InMemoryKVStore) Set(key kv.Key, value []byte) {
	if value == nil {
		b.Del(key)
		return
	}

	if b.marking {
		b.uncommitedMarked[key] = struct{}{}
	}

	b.uncommitted.Set(key, value)
}

func (b *InMemoryKVStore) Del(key kv.Key) {
	if b.marking {
		delete(b.uncommitedMarked, key)
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
