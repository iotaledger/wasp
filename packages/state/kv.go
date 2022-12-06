// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package state

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/trie"
)

// trieKVAdapter is a KVStoreReader backed by a TrieReader
type trieKVAdapter struct {
	*trie.TrieReader
}

var _ kv.KVStoreReader = &trieKVAdapter{}

func (t *trieKVAdapter) Get(key kv.Key) ([]byte, error) {
	return t.TrieReader.Get([]byte(key)), nil
}

func (t *trieKVAdapter) Has(key kv.Key) (bool, error) {
	return t.TrieReader.Has([]byte(key)), nil
}

func (t *trieKVAdapter) Iterate(prefix kv.Key, f func(kv.Key, []byte) bool) error {
	t.TrieReader.Iterator([]byte(prefix)).Iterate(func(k []byte, v []byte) bool {
		return f(kv.Key(k), v)
	})
	return nil
}

func (t *trieKVAdapter) IterateKeys(prefix kv.Key, f func(kv.Key) bool) error {
	t.TrieReader.Iterator([]byte(prefix)).IterateKeys(func(k []byte) bool {
		return f(kv.Key(k))
	})
	return nil
}

func (t *trieKVAdapter) IterateKeysSorted(prefix kv.Key, f func(key kv.Key) bool) error {
	return t.IterateKeys(prefix, f)
}

func (t *trieKVAdapter) IterateSorted(prefix kv.Key, f func(key kv.Key, value []byte) bool) error {
	return t.Iterate(prefix, f)
}

func (t *trieKVAdapter) MustGet(key kv.Key) []byte {
	return t.TrieReader.Get([]byte(key))
}

func (t *trieKVAdapter) MustHas(key kv.Key) bool {
	return t.TrieReader.Has([]byte(key))
}

func (t *trieKVAdapter) MustIterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	kv.MustIterate(t, prefix, f)
}

func (t *trieKVAdapter) MustIterateKeys(prefix kv.Key, f func(key kv.Key) bool) {
	kv.MustIterateKeys(t, prefix, f)
}

func (t *trieKVAdapter) MustIterateKeysSorted(prefix kv.Key, f func(key kv.Key) bool) {
	kv.MustIterateKeysSorted(t, prefix, f)
}

func (t *trieKVAdapter) MustIterateSorted(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	kv.MustIterateSorted(t, prefix, f)
}
