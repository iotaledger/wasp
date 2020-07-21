package kv

import (
	"encoding/hex"
	"fmt"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/mr-tron/base58"
)

// BufferedKVStore represents a KVStore backed by a database. Writes are cached in-memory as
// a MutationSequence; reads are delegated to the backing database when not cached.
type BufferedKVStore interface {
	KVStore

	// the uncommitted mutations
	Mutations() MutationSequence
	ClearMutations()
	Clone() BufferedKVStore

	Codec() Codec

	// only for testing!
	DangerouslyDumpToMap() Map
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
	mutations MutationSequence
}

func NewBufferedKVStore(db kvstore.KVStore) BufferedKVStore {
	return &bufferedKVStore{
		db:        db,
		mutations: NewMutationSequence(),
	}
}

func (b *bufferedKVStore) Clone() BufferedKVStore {
	return &bufferedKVStore{
		db:        b.db,
		mutations: b.mutations.Clone(),
	}
}

func (b *bufferedKVStore) Codec() Codec {
	return NewCodec(b)
}

func (b *bufferedKVStore) Mutations() MutationSequence {
	return b.mutations
}

func (b *bufferedKVStore) ClearMutations() {
	b.mutations = NewMutationSequence()
}

// iterates over all key-value pairs in KVStore
func (b *bufferedKVStore) DangerouslyDumpToMap() Map {
	prefix := len(b.db.Realm())
	ret := NewMap()
	err := b.db.Iterate(kvstore.EmptyPrefix, func(key kvstore.Key, value kvstore.Value) bool {
		ret.Set(Key(key[prefix:]), value)
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
	for k, v := range b.DangerouslyDumpToMap().ToGoMap() {
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

func (b *bufferedKVStore) flag(k Key) string {
	mut := b.mutations.Latest(k)
	if mut != nil {
		return "+"
	}
	return " "
}

func (b *bufferedKVStore) Set(key Key, value []byte) {
	b.mutations.Add(NewMutationSet(key, value))
}

func (b *bufferedKVStore) Del(key Key) {
	b.mutations.Add(NewMutationDel(key))
}

func (b *bufferedKVStore) Get(key Key) ([]byte, error) {
	mut := b.mutations.Latest(key)
	if mut != nil {
		return mut.Value(), nil
	}
	v, err := b.db.Get(kvstore.Key(key))
	if err == kvstore.ErrKeyNotFound {
		return nil, nil
	}
	return v, asDBError(err)
}

func (b *bufferedKVStore) Has(key Key) (bool, error) {
	mut := b.mutations.Latest(key)
	if mut != nil {
		return mut.Value() != nil, nil
	}
	v, err := b.db.Has(kvstore.Key(key))
	return v, asDBError(err)
}

func (b *bufferedKVStore) Iterate(prefix Key, f func(key Key, value []byte) bool) error {
	seen, done := b.mutations.IterateValues(prefix, f)
	if done {
		return nil
	}
	realm := len(b.db.Realm())
	return b.db.Iterate([]byte(prefix), func(key kvstore.Key, value kvstore.Value) bool {
		k := Key(key[realm:])
		_, ok := seen[k]
		if ok {
			return true
		}
		return f(k, value)
	})
}

func (b *bufferedKVStore) IterateKeys(prefix Key, f func(key Key) bool) error {
	seen, done := b.mutations.IterateValues(prefix, func(key Key, value []byte) bool {
		return f(key)
	})
	if done {
		return nil
	}
	realm := len(b.db.Realm())
	return b.db.IterateKeys([]byte(prefix), func(key kvstore.Key) bool {
		k := Key(key[realm:])
		_, ok := seen[k]
		if ok {
			return true
		}
		return f(k)
	})
}
