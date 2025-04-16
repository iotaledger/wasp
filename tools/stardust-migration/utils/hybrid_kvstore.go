package utils

import (
	"bytes"

	"github.com/iotaledger/hive.go/kvstore"
)

func NewHybridKVStore(defaultStore kvstore.KVStore, storesByPrefix map[string]kvstore.KVStore) *HybridKVStore {
	return &HybridKVStore{
		def:            defaultStore,
		storesByPrefix: storesByPrefix,
		initialKeys:    make(map[string]bool),
	}
}

type HybridKVStore struct {
	def            kvstore.KVStore
	storesByPrefix map[string]kvstore.KVStore // We could add trie here for faster prefix determination
	initialKeys    map[string]bool
}

var _ kvstore.KVStore = &HybridKVStore{}

func (h *HybridKVStore) LoadAllFromDefault() (int, error) {
	count := 0

	for prefix, store := range h.storesByPrefix {
		h.def.Iterate([]byte(prefix), func(key kvstore.Key, value kvstore.Value) bool {
			if err := store.Set(key, value); err != nil {
				panic(err)
			}
			h.initialKeys[string(key)] = true
			count++
			return true
		})
	}

	return count, nil
}

func (h *HybridKVStore) CopyAllToDefault() (int, error) {
	def, err := h.def.Batched()
	if err != nil {
		return 0, err
	}

	count := 0

	for _, store := range h.storesByPrefix {
		var err error

		store.Iterate(nil, func(key kvstore.Key, value kvstore.Value) bool {
			if err = def.Set(key, value); err != nil {
				return false
			}
			h.initialKeys[string(key)] = false
			count++
			return true
		})
		if err != nil {
			return 0, err
		}
	}

	for key, wasDeleted := range h.initialKeys {
		if wasDeleted {
			if err := def.Delete([]byte(key)); err != nil {
				return 0, err
			}
			count++
		}
	}

	if err := def.Commit(); err != nil {
		return 0, err
	}

	return count, nil
}

func (h *HybridKVStore) WithRealm(realm kvstore.Realm) (kvstore.KVStore, error) {
	panic("not implemented")
}

func (h *HybridKVStore) WithExtendedRealm(realm kvstore.Realm) (kvstore.KVStore, error) {
	panic("not implemented")
}

func (h *HybridKVStore) Realm() kvstore.Realm {
	panic("not implemented")
}

func (h *HybridKVStore) Iterate(prefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyValueConsumerFunc, direction ...kvstore.IterDirection) error {
	for storePrefix, store := range h.storesByPrefix {
		if bytes.HasPrefix(prefix, []byte(storePrefix)) {
			return store.Iterate(prefix, consumerFunc, direction...)
		}
	}

	return h.def.Iterate(prefix, consumerFunc, direction...)
}

func (h *HybridKVStore) IterateKeys(prefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyConsumerFunc, direction ...kvstore.IterDirection) error {
	for storePrefix, store := range h.storesByPrefix {
		if bytes.HasPrefix(prefix, []byte(storePrefix)) {
			return store.IterateKeys(prefix, consumerFunc, direction...)
		}
	}

	return h.def.IterateKeys(prefix, consumerFunc, direction...)
}

func (h *HybridKVStore) Clear() error {
	if err := h.def.Clear(); err != nil {
		return err
	}

	for _, store := range h.storesByPrefix {
		if err := store.Clear(); err != nil {
			return err
		}
	}

	return nil
}

func (h *HybridKVStore) Get(key kvstore.Key) (kvstore.Value, error) {
	for prefix, store := range h.storesByPrefix {
		if bytes.HasPrefix(key, []byte(prefix)) {
			return store.Get(key)
		}
	}

	return h.def.Get(key)
}

func (h *HybridKVStore) Set(key kvstore.Key, value kvstore.Value) error {
	for prefix, store := range h.storesByPrefix {
		if bytes.HasPrefix(key, []byte(prefix)) {
			return store.Set(key, value)
		}
	}

	return h.def.Set(key, value)
}

func (h *HybridKVStore) Has(key kvstore.Key) (bool, error) {
	for prefix, store := range h.storesByPrefix {
		if bytes.HasPrefix(key, []byte(prefix)) {
			return store.Has(key)
		}
	}

	return h.def.Has(key)
}

func (h *HybridKVStore) Delete(key kvstore.Key) error {
	for prefix, store := range h.storesByPrefix {
		if bytes.HasPrefix(key, []byte(prefix)) {
			return store.Delete(key)
		}
	}

	return h.def.Delete(key)
}

func (h *HybridKVStore) DeletePrefix(prefix kvstore.KeyPrefix) error {
	for storePrefix, store := range h.storesByPrefix {
		if bytes.HasPrefix(prefix, []byte(storePrefix)) {
			return store.DeletePrefix(prefix)
		}
	}

	return h.def.DeletePrefix(prefix)
}

func (h *HybridKVStore) Flush() error {
	if err := h.def.Flush(); err != nil {
		return err
	}

	for _, store := range h.storesByPrefix {
		if err := store.Flush(); err != nil {
			return err
		}
	}

	return nil
}

func (h *HybridKVStore) Close() error {
	if err := h.def.Close(); err != nil {
		return err
	}

	for _, store := range h.storesByPrefix {
		if err := store.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (h *HybridKVStore) Batched() (kvstore.BatchedMutations, error) {
	return newBatchedHybridKVStore(h)
}

func newBatchedHybridKVStore(s *HybridKVStore) (*batchedHybridKVStore, error) {
	def, err := s.def.Batched()
	if err != nil {
		return nil, err
	}

	storesByPrefix := make(map[string]kvstore.BatchedMutations, len(s.storesByPrefix))

	for prefix, store := range s.storesByPrefix {
		var err error
		storesByPrefix[prefix], err = store.Batched()
		if err != nil {
			return nil, err
		}
	}

	return &batchedHybridKVStore{
		def:            def,
		storesByPrefix: storesByPrefix,
	}, nil
}

type batchedHybridKVStore struct {
	def            kvstore.BatchedMutations
	storesByPrefix map[string]kvstore.BatchedMutations
}

func (b *batchedHybridKVStore) Set(key kvstore.Key, value kvstore.Value) error {
	for prefix, store := range b.storesByPrefix {
		if bytes.HasPrefix(key, []byte(prefix)) {
			return store.Set(key, value)
		}
	}

	return b.def.Set(key, value)
}

func (b *batchedHybridKVStore) Delete(key kvstore.Key) error {
	for prefix, store := range b.storesByPrefix {
		if bytes.HasPrefix(key, []byte(prefix)) {
			return store.Delete(key)
		}
	}

	return b.def.Delete(key)
}

func (b *batchedHybridKVStore) Cancel() {
	b.def.Cancel()

	for _, store := range b.storesByPrefix {
		store.Cancel()
	}
}

func (b *batchedHybridKVStore) Commit() error {
	if err := b.def.Commit(); err != nil {
		return err
	}

	for _, store := range b.storesByPrefix {
		if err := store.Commit(); err != nil {
			return err
		}
	}

	return nil
}
