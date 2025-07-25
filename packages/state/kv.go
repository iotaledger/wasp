// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package state

import (
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/trie"
)

// trieKVAdapter is a KVStoreReader backed by a TrieReader
type trieKVAdapter struct {
	*trie.TrieReader
}

var _ kv.KVStoreReader = &trieKVAdapter{}

func (t *trieKVAdapter) Get(key kv.Key) []byte {
	return t.TrieReader.Get([]byte(key))
}

func (t *trieKVAdapter) Has(key kv.Key) bool {
	return t.TrieReader.Has([]byte(key))
}

func (t *trieKVAdapter) Iterate(prefix kv.Key, f func(kv.Key, []byte) bool) {
	t.TrieReader.Iterator([]byte(prefix)).Iterate(func(k []byte, v []byte) bool {
		return f(kv.Key(k), v)
	})
}

func (t *trieKVAdapter) IterateKeys(prefix kv.Key, f func(kv.Key) bool) {
	t.TrieReader.Iterator([]byte(prefix)).IterateKeys(func(k []byte) bool {
		return f(kv.Key(k))
	})
}

func (t *trieKVAdapter) IterateKeysSorted(prefix kv.Key, f func(key kv.Key) bool) {
	t.IterateKeys(prefix, f)
}

func (t *trieKVAdapter) IterateSorted(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	t.Iterate(prefix, f)
}
