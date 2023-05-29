package trie_test

import (
	"bytes"

	"github.com/iotaledger/wasp/packages/trie"
)

// ----------------------------------------------------------------------------
// InMemoryKVStore is a KVStore implementation. Mostly used for testing
var (
	_ trie.KVStore     = InMemoryKVStore{}
	_ trie.Traversable = InMemoryKVStore{}
	_ trie.KVIterator  = &simpleInMemoryIterator{}
)

type (
	InMemoryKVStore map[string][]byte

	simpleInMemoryIterator struct {
		store  InMemoryKVStore
		prefix []byte
	}
)

func NewInMemoryKVStore() InMemoryKVStore {
	return make(InMemoryKVStore)
}

func (im InMemoryKVStore) Get(k []byte) []byte {
	return im[string(k)]
}

func (im InMemoryKVStore) Has(k []byte) bool {
	_, ok := im[string(k)]
	return ok
}

func (im InMemoryKVStore) Iterate(f func(k []byte, v []byte) bool) {
	for k, v := range im {
		if !f([]byte(k), v) {
			return
		}
	}
}

func (im InMemoryKVStore) IterateKeys(f func(k []byte) bool) {
	for k := range im {
		if !f([]byte(k)) {
			return
		}
	}
}

func (im InMemoryKVStore) Set(k, v []byte) {
	if len(v) != 0 {
		im[string(k)] = v
	} else {
		delete(im, string(k))
	}
}

func (im InMemoryKVStore) Del(k []byte) {
	delete(im, string(k))
}

func (im InMemoryKVStore) Iterator(prefix []byte) trie.KVIterator {
	return &simpleInMemoryIterator{
		store:  im,
		prefix: prefix,
	}
}

func (si *simpleInMemoryIterator) Iterate(f func(k []byte, v []byte) bool) {
	var key []byte
	for k, v := range si.store {
		key = []byte(k)
		if bytes.HasPrefix(key, si.prefix) {
			if !f(key, v) {
				return
			}
		}
	}
}

func (si *simpleInMemoryIterator) IterateKeys(f func(k []byte) bool) {
	var key []byte
	for k := range si.store {
		key = []byte(k)
		if bytes.HasPrefix(key, si.prefix) {
			if !f(key) {
				return
			}
		}
	}
}
