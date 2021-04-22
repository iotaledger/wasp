package buffered

import (
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/mr-tron/base58"
)

// BufferedKVStore represents a KVStore backed by a database. Writes are cached in-memory;
// reads are delegated to the backing database when not cached.
type BufferedKVStore interface {
	kv.KVStore

	// the uncommitted mutations
	Mutations() *Mutations
	ClearMutations()
	Clone() BufferedKVStore

	// only for testing!
	DangerouslyDumpToDict() dict.Dict
	// only for testing!
	DangerouslyDumpToString() string
}

// Any failed access to the DB will return an instance of DBError
type DBError struct{ error }

func (d DBError) Error() string {
	return d.error.Error()
}

func asDBError(e error) error {
	if e == nil {
		return nil
	}
	if d, ok := e.(DBError); ok {
		return d
	}
	return DBError{e}
}

type bufferedKVStore struct {
	db        kvstore.KVStore
	mutations *Mutations
}

func NewBufferedKVStore(db kvstore.KVStore) BufferedKVStore {
	return &bufferedKVStore{
		db:        db,
		mutations: NewMutations(),
	}
}

func (b *bufferedKVStore) Clone() BufferedKVStore {
	return &bufferedKVStore{
		db:        b.db,
		mutations: b.mutations.Clone(),
	}
}

func (b *bufferedKVStore) Mutations() *Mutations {
	return b.mutations
}

func (b *bufferedKVStore) ClearMutations() {
	b.mutations = NewMutations()
}

// iterates over all key-value pairs in KVStore
func (b *bufferedKVStore) DangerouslyDumpToDict() dict.Dict {
	ret := dict.New()
	err := b.db.Iterate(kvstore.EmptyPrefix, func(key kvstore.Key, value kvstore.Value) bool {
		ret.Set(kv.Key(key), value)
		return true
	})
	if err != nil {
		panic(asDBError(err))
	}
	b.mutations.ApplyTo(ret)
	return ret
}

// iterates over all key-value pairs in KVStore
func (b *bufferedKVStore) DangerouslyDumpToString() string {
	ret := "         BufferedKVStore:\n"
	for k, v := range b.DangerouslyDumpToDict() {
		ret += fmt.Sprintf(
			"           [%s] 0x%s: 0x%s (base58: %s)\n",
			b.flag(k),
			slice(hex.EncodeToString([]byte(k))),
			slice(hex.EncodeToString(v)),
			slice(base58.Encode(v)),
		)
	}
	return ret
}

func slice(s string) string {
	if len(s) > 44 {
		return s[:10] + "[...]" + s[len(s)-10:]
	}
	return s
}

func (b *bufferedKVStore) flag(k kv.Key) string {
	if b.mutations.Contains(k) {
		return "+"
	}
	return " "
}

func (b *bufferedKVStore) Set(key kv.Key, value []byte) {
	b.mutations.Set(key, value)
}

func (b *bufferedKVStore) Del(key kv.Key) {
	b.mutations.Del(key)
}

func (b *bufferedKVStore) Get(key kv.Key) ([]byte, error) {
	v, ok := b.mutations.Get(key)
	if ok {
		return v, nil
	}
	v, err := b.db.Get(kvstore.Key(key))
	if err == kvstore.ErrKeyNotFound {
		return nil, nil
	}
	return v, asDBError(err)
}

func (b *bufferedKVStore) MustGet(key kv.Key) []byte {
	return kv.MustGet(b, key)
}

func (b *bufferedKVStore) Has(key kv.Key) (bool, error) {
	v, ok := b.mutations.Get(key)
	if ok {
		return v != nil, nil
	}
	ok, err := b.db.Has(kvstore.Key(key))
	return ok, asDBError(err)
}

func (b *bufferedKVStore) MustHas(key kv.Key) bool {
	return kv.MustHas(b, key)
}

func (b *bufferedKVStore) Iterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) error {
	var err error
	err2 := b.IterateKeys(prefix, func(k kv.Key) bool {
		var v []byte
		v, err = b.Get(k)
		if err != nil {
			return false
		}
		return f(k, v)
	})
	if err2 != nil {
		return err2
	}
	return err
}

func (b *bufferedKVStore) MustIterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	kv.MustIterate(b, prefix, f)
}

func (b *bufferedKVStore) IterateKeys(prefix kv.Key, f func(key kv.Key) bool) error {
	for k := range b.mutations.Sets {
		if !k.HasPrefix(prefix) {
			continue
		}
		if !f(k) {
			return nil
		}
	}
	return b.db.IterateKeys([]byte(prefix), func(k kvstore.Key) bool {
		{
			k := kv.Key(k)
			if !b.mutations.Contains(k) {
				return f(k)
			}
		}
		return true
	})
}

func (b *bufferedKVStore) IterateSorted(prefix kv.Key, f func(key kv.Key, value []byte) bool) error {
	var err error
	err2 := b.IterateKeysSorted(prefix, func(k kv.Key) bool {
		var v []byte
		v, err = b.Get(k)
		if err != nil {
			return false
		}
		return f(k, v)
	})
	if err2 != nil {
		return err2
	}
	return err
}

func (b *bufferedKVStore) MustIterateSorted(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	kv.MustIterateSorted(b, prefix, f)
}

func (b *bufferedKVStore) IterateKeysSorted(prefix kv.Key, f func(key kv.Key) bool) error {
	var keys []kv.Key

	for k := range b.mutations.Sets {
		if !k.HasPrefix(prefix) {
			continue
		}
		keys = append(keys, k)
	}

	err := b.db.IterateKeys([]byte(prefix), func(k kvstore.Key) bool {
		{
			k := kv.Key(k)
			if !b.mutations.Contains(k) {
				keys = append(keys, k)
			}
		}
		return true
	})
	if err != nil {
		return err
	}

	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	for _, k := range keys {
		if !f(k) {
			break
		}
	}
	return nil
}

func (b *bufferedKVStore) MustIterateKeys(prefix kv.Key, f func(key kv.Key) bool) {
	kv.MustIterateKeys(b, prefix, f)
}

func (b *bufferedKVStore) MustIterateKeysSorted(prefix kv.Key, f func(key kv.Key) bool) {
	kv.MustIterateKeysSorted(b, prefix, f)
}
