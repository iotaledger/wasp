package trie

import (
	"errors"

	"github.com/iotaledger/hive.go/core/kvstore"
)

// HiveKVStoreAdapter maps a partition of the Hive KVStore to trie_go.KVStore
type HiveKVStoreAdapter struct {
	kvs    kvstore.KVStore
	prefix []byte
}

// NewHiveKVStoreAdapter creates a new KVStore as a partition of hive.go KVStore
func NewHiveKVStoreAdapter(kvs kvstore.KVStore, prefix []byte) *HiveKVStoreAdapter {
	return &HiveKVStoreAdapter{kvs: kvs, prefix: prefix}
}

func mustNoErr(err error) {
	if err != nil {
		panic(err)
	}
}

func makeKey(prefix, k []byte) []byte {
	if len(prefix) == 0 {
		return k
	}
	return concat(prefix, k)
}

func (kvs *HiveKVStoreAdapter) Get(key []byte) []byte {
	v, err := kvs.kvs.Get(makeKey(kvs.prefix, key))
	if errors.Is(err, kvstore.ErrKeyNotFound) {
		return nil
	}
	mustNoErr(err)
	if len(v) == 0 {
		return nil
	}
	return v
}

func (kvs *HiveKVStoreAdapter) Has(key []byte) bool {
	v, err := kvs.kvs.Get(makeKey(kvs.prefix, key))
	if errors.Is(err, kvstore.ErrKeyNotFound) || len(v) == 0 {
		return false
	}
	mustNoErr(err)
	return true
}

func (kvs *HiveKVStoreAdapter) Set(key, value []byte) {
	var err error
	if len(value) == 0 {
		err = kvs.kvs.Delete(makeKey(kvs.prefix, key))
	} else {
		err = kvs.kvs.Set(makeKey(kvs.prefix, key), value)
	}
	mustNoErr(err)
}

func (kvs *HiveKVStoreAdapter) Iterate(fun func(k []byte, v []byte) bool) {
	err := kvs.kvs.Iterate(kvs.prefix, func(key kvstore.Key, value kvstore.Value) bool {
		return fun(key[len(kvs.prefix):], value)
	})
	mustNoErr(err)
}

func (kvs *HiveKVStoreAdapter) IterateKeys(fun func(k []byte) bool) {
	err := kvs.kvs.IterateKeys(kvs.prefix, func(key kvstore.Key) bool {
		return fun(key[len(kvs.prefix):])
	})
	mustNoErr(err)
}
