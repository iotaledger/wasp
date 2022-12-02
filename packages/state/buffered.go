// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package state

import (
	"github.com/iotaledger/hive.go/core/kvstore"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
)

// bufferedKVStore is a KVStore backed by some other KVStore. Writes are cached in-memory;
// reads are delegated to the backing store for unmodified keys.
type bufferedKVStore struct {
	db   kvstore.KVStore
	muts *buffered.Mutations
}

var _ kvstore.KVStore = &bufferedKVStore{}

func newBufferedKVStore(db kvstore.KVStore) *bufferedKVStore {
	return &bufferedKVStore{
		db:   db,
		muts: buffered.NewMutations(),
	}
}

func (*bufferedKVStore) Batched() (kvstore.BatchedMutations, error) {
	panic("should no be called")
}

func (*bufferedKVStore) Clear() error {
	panic("should no be called")
}

func (*bufferedKVStore) Close() error {
	panic("should no be called")
}

func (b *bufferedKVStore) Delete(key []byte) error {
	b.muts.Del(kv.Key(key))
	return nil
}

func (*bufferedKVStore) DeletePrefix(prefix []byte) error {
	panic("should no be called")
}

func (*bufferedKVStore) Flush() error {
	panic("should no be called")
}

func (b *bufferedKVStore) Get(key []byte) (value []byte, err error) {
	v, ok := b.muts.Get(kv.Key(key))
	if ok {
		return v, nil
	}
	return b.db.Get(key)
}

func (b *bufferedKVStore) Has(key []byte) (bool, error) {
	v, ok := b.muts.Get(kv.Key(key))
	if ok {
		return v != nil, nil
	}
	return b.db.Has(key)
}

func (*bufferedKVStore) Iterate(prefix []byte, kvConsumerFunc kvstore.IteratorKeyValueConsumerFunc, direction ...kvstore.IterDirection) error {
	panic("should no be called")
}

func (*bufferedKVStore) IterateKeys(prefix []byte, consumerFunc kvstore.IteratorKeyConsumerFunc, direction ...kvstore.IterDirection) error {
	panic("should no be called")
}

func (b *bufferedKVStore) Realm() []byte {
	return b.db.Realm()
}

func (b *bufferedKVStore) Set(key []byte, value []byte) error {
	b.muts.Set(kv.Key(key), value)
	return nil
}

func (*bufferedKVStore) WithRealm(realm []byte) (kvstore.KVStore, error) {
	panic("should no be called")
}
