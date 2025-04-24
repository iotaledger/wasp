package utils

import (
	"bytes"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

func NewInMemoryKVStore(readOnlyUncommitted bool) *InMemoryKVStore {
	committed := dict.New()

	var uncommitted *buffered.BufferedKVStore
	if readOnlyUncommitted {
		uncommitted = buffered.NewBufferedKVStore(NoopKVStoreReader[kv.Key]{})
	} else {
		uncommitted = buffered.NewBufferedKVStore(committed)
	}

	return &InMemoryKVStore{
		committed:           committed,
		uncommitted:         uncommitted,
		committedMarked:     make(map[kv.Key]struct{}),
		uncommitedMarked:    make(map[kv.Key]struct{}),
		readOnlyUncommitted: readOnlyUncommitted,
	}
}

type InMemoryKVStore struct {
	committed           dict.Dict
	uncommitted         *buffered.BufferedKVStore
	marking             bool
	uncommitedMarked    map[kv.Key]struct{}
	committedMarked     map[kv.Key]struct{}
	prevMutsCount       int
	readOnlyUncommitted bool
	keyValidator        func(k kv.Key, v []byte)
}

var _ kv.KVStoreReader = &InMemoryKVStore{}
var _ kv.KVStore = &InMemoryKVStore{}

func (b *InMemoryKVStore) Len() int {
	return b.CommittedSize()
}

func (b *InMemoryKVStore) CommittedSize() int {
	return len(b.committed)
}

func (b *InMemoryKVStore) CommittedState() (state map[kv.Key][]byte, marks map[kv.Key]struct{}) {
	return b.committed, b.committedMarked
}

func (b *InMemoryKVStore) SetCommittedState(state map[kv.Key][]byte, marks map[kv.Key]struct{}) {
	if state != nil {
		b.committed = state

		if b.readOnlyUncommitted {
			b.uncommitted = buffered.NewBufferedKVStore(NoopKVStoreReader[kv.Key]{})
		} else {
			b.uncommitted = buffered.NewBufferedKVStore(b.committed)
		}
	}
	if marks != nil {
		b.committedMarked = marks
	}
}

func (b *InMemoryKVStore) SetKeyValidator(keyValidator func(k kv.Key, v []byte)) {
	b.keyValidator = keyValidator
}

func (b *InMemoryKVStore) Mutations() *buffered.Mutations {
	return b.uncommitted.Mutations()
}

func (b *InMemoryKVStore) MutationsCount() int {
	return len(b.uncommitted.Mutations().Sets) + len(b.uncommitted.Mutations().Dels)
}

func (b *InMemoryKVStore) MutationsCountDiff() int {
	newMutsCount := b.MutationsCount()
	mutsCountDiff := newMutsCount - b.prevMutsCount
	b.prevMutsCount = newMutsCount
	return mutsCountDiff
}

func (b *InMemoryKVStore) StartMarking(readOnlyUncommitted bool) {
	b.marking = true
	if readOnlyUncommitted {
		b.uncommitted.SetKVStoreReader(NoopKVStoreReader[kv.Key]{})
	} else {
		b.uncommitted.SetKVStoreReader(b.committed)
	}
}

func (b *InMemoryKVStore) StopMarking() {
	b.marking = false
	if b.readOnlyUncommitted {
		b.uncommitted.SetKVStoreReader(NoopKVStoreReader[kv.Key]{})
	} else {
		b.uncommitted.SetKVStoreReader(b.committed)
	}
}

// DeleteIfNotSet deletes marked entries if they where not set since the last commit
func (b *InMemoryKVStore) DeleteMarkedIfNotSet() int {
	uncommittedSets := b.uncommitted.Mutations().Sets

	count := 0
	for key := range b.committedMarked {
		if _, isSet := uncommittedSets[key]; !isSet {
			b.uncommitted.Del(key)
			count++
		}
	}

	return count
}

// DeleteIfNotSet deletes entries if they where not set since the last commit
func (b *InMemoryKVStore) DeleteIfNotSet() {
	uncommittedSets := b.uncommitted.Mutations().Sets

	for key := range b.committed {
		if _, isSet := uncommittedSets[key]; !isSet {
			b.uncommitted.Del(key)
		}
	}
}

// RemoveRedundantMutations removes mutations that have no effect
func (b *InMemoryKVStore) RemoveRedundantMutations() int {
	committedSets := b.committed
	uncommittedSets := b.uncommitted.Mutations().Sets
	removedCount := 0

	for key, value := range uncommittedSets {
		committed, ok := committedSets[key]
		if ok && bytes.Equal(committed, value) {
			removedCount++
			delete(uncommittedSets, key)
		}
	}

	uncommittedDels := b.uncommitted.Mutations().Dels

	for key := range uncommittedDels {
		_, ok := committedSets[key]
		if !ok {
			removedCount++
			delete(uncommittedDels, key)
		}
	}

	return removedCount
}

func (b *InMemoryKVStore) Commit(onlyEffectiveMutations bool) (*buffered.Mutations, int) {
	var ineffectiveRemoved int
	if onlyEffectiveMutations {
		ineffectiveRemoved = b.RemoveRedundantMutations()
	}

	muts := b.uncommitted.Mutations()
	muts.ApplyTo(b.committed)
	b.uncommitted.SetMutations(buffered.NewMutations())
	b.committedMarked = b.uncommitedMarked
	b.uncommitedMarked = make(map[kv.Key]struct{}, len(b.uncommitedMarked))

	b.prevMutsCount = 0

	return muts, ineffectiveRemoved
}

func (b *InMemoryKVStore) Set(key kv.Key, value []byte) {
	if value == nil {
		b.Del(key)
		return
	}

	if b.keyValidator != nil {
		b.keyValidator(key, value)
	}

	if b.marking {
		b.uncommitedMarked[key] = struct{}{}
	}

	b.uncommitted.Set(key, value)
}

func (b *InMemoryKVStore) Del(key kv.Key) {
	if b.keyValidator != nil {
		b.keyValidator(key, nil)
	}

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
