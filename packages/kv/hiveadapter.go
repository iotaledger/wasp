package kv

import (
	"errors"
	"sort"

	"github.com/iotaledger/hive.go/core/kvstore"
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
func (h *HiveKVStoreReader) Get(key Key) ([]byte, error) {
	b, err := h.db.Get(kvstore.Key(key))
	if errors.Is(err, kvstore.ErrKeyNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, wrapDBError(err)
	}
	return wrapBytes(b), nil
}

func (h *HiveKVStoreReader) Has(key Key) (bool, error) {
	ok, err := h.db.Has(kvstore.Key(key))
	return ok, wrapDBError(err)
}

func (h *HiveKVStoreReader) Iterate(prefix Key, f func(key Key, value []byte) bool) error {
	return wrapDBError(h.db.Iterate([]byte(prefix), func(k kvstore.Key, v kvstore.Value) bool {
		return f(Key(k), wrapBytes(v))
	}))
}

func (h *HiveKVStoreReader) IterateKeys(prefix Key, f func(key Key) bool) error {
	return wrapDBError(h.db.IterateKeys([]byte(prefix), func(k kvstore.Key) bool {
		return f(Key(k))
	}))
}

func (h *HiveKVStoreReader) IterateSorted(prefix Key, f func(key Key, value []byte) bool) error {
	var err error
	err2 := h.IterateKeysSorted(prefix, func(k Key) bool {
		var v []byte
		v, err = h.Get(k)
		if err != nil {
			return false
		}
		return f(k, wrapBytes(v))
	})
	if err2 != nil {
		return err2
	}
	return wrapDBError(err)
}

func (h *HiveKVStoreReader) IterateKeysSorted(prefix Key, f func(key Key) bool) error {
	var keys []Key
	err := h.db.IterateKeys([]byte(prefix), func(k kvstore.Key) bool {
		keys = append(keys, Key(k))
		return true
	})
	if err != nil {
		return wrapDBError(err)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	for _, k := range keys {
		if !f(k) {
			break
		}
	}
	return nil
}

// MustGet returns the value, or nil if not found
func (h *HiveKVStoreReader) MustGet(key Key) []byte {
	return MustGet(h, key)
}

func (h *HiveKVStoreReader) MustHas(key Key) bool {
	return MustHas(h, key)
}

func (h *HiveKVStoreReader) MustIterate(prefix Key, f func(key Key, value []byte) bool) {
	MustIterate(h, prefix, f)
}

func (h *HiveKVStoreReader) MustIterateKeys(prefix Key, f func(key Key) bool) {
	MustIterateKeys(h, prefix, f)
}

func (h *HiveKVStoreReader) MustIterateSorted(prefix Key, f func(key Key, value []byte) bool) {
	MustIterateSorted(h, prefix, f)
}

func (h *HiveKVStoreReader) MustIterateKeysSorted(prefix Key, f func(key Key) bool) {
	MustIterateKeysSorted(h, prefix, f)
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

func wrapDBError(e error) error {
	if e == nil {
		return nil
	}
	if d, ok := e.(*DBError); ok {
		return d
	}
	return &DBError{e}
}
