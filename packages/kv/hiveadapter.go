package kv

import (
	"errors"
	"sort"

	"github.com/iotaledger/hive.go/kvstore"
)

// HiveKVStoreReader is an implementation of KVStoreReader with an instance of
// hive's kvstore.KVStore as backend.
type HiveKVStoreReader struct {
	db kvstore.KVStore
}

var _ KVStoreReader = &HiveKVStoreReader{}

func NewHiveKVStoreReader(db kvstore.KVStore) *HiveKVStoreReader {
	return &HiveKVStoreReader{db}
}

// Get returns the value, or nil if not found
func (h *HiveKVStoreReader) Get(key Key) []byte {
	b, err := h.db.Get(kvstore.Key(key))
	if errors.Is(err, kvstore.ErrKeyNotFound) {
		return nil
	}
	checkDBError(err)
	return wrapBytes(b)
}

func (h *HiveKVStoreReader) Has(key Key) bool {
	ok, err := h.db.Has(kvstore.Key(key))
	checkDBError(err)
	return ok
}

func (h *HiveKVStoreReader) Iterate(prefix Key, f func(key Key, value []byte) bool) {
	err := h.db.Iterate([]byte(prefix), func(k kvstore.Key, v kvstore.Value) bool {
		return f(Key(k), wrapBytes(v))
	})
	checkDBError(err)
}

func (h *HiveKVStoreReader) IterateKeys(prefix Key, f func(key Key) bool) {
	err := h.db.IterateKeys([]byte(prefix), func(k kvstore.Key) bool {
		return f(Key(k))
	})
	checkDBError(err)
}

func (h *HiveKVStoreReader) IterateSorted(prefix Key, f func(key Key, value []byte) bool) {
	h.IterateKeysSorted(prefix, func(k Key) bool {
		return f(k, wrapBytes(h.Get(k)))
	})
}

func (h *HiveKVStoreReader) IterateKeysSorted(prefix Key, f func(key Key) bool) {
	var keys []Key
	err := h.db.IterateKeys([]byte(prefix), func(k kvstore.Key) bool {
		keys = append(keys, Key(k))
		return true
	})
	checkDBError(err)
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	for _, k := range keys {
		if !f(k) {
			break
		}
	}
}

type DBError struct{ error }

func (d *DBError) Error() string {
	return d.error.Error()
}

func (d *DBError) Unwrap() error {
	return d.error
}

func wrapBytes(b []byte) []byte {
	// after Set(k, []byte{}), make sure we return []byte{} instead of nil
	if b == nil {
		return []byte{}
	}
	return b
}

func checkDBError(e error) {
	if e == nil {
		return
	}
	if d, ok := e.(*DBError); ok {
		panic(d)
	}
	panic(&DBError{e})
}
