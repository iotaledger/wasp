package kv

import (
	"encoding/hex"
	"fmt"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/mr-tron/base58"
)

// BufferedKVStore represents a KVStore backed by a database. Writes are cached in-memory as
// a MutationMap; reads are delegated to the backing database when not cached.
type BufferedKVStore interface {
	KVStore

	// the uncommitted mutations
	Mutations() MutationMap
	ClearMutations()
	Clone() BufferedKVStore

	Codec() Codec

	// only for testing!
	DangerouslyDumpToMap() map[Key][]byte
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
	mutations MutationMap
}

func NewBufferedKVStore(db kvstore.KVStore) BufferedKVStore {
	return &bufferedKVStore{
		db:        db,
		mutations: NewMutationMap(),
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

func (b *bufferedKVStore) Mutations() MutationMap {
	return b.mutations
}

func (b *bufferedKVStore) ClearMutations() {
	b.mutations = NewMutationMap()
}

// iterates over all key-value pairs in KVStore
func (b *bufferedKVStore) DangerouslyDumpToMap() map[Key][]byte {
	prefix := len(b.db.Realm())
	ret := make(map[Key][]byte)
	err := b.db.Iterate(kvstore.EmptyPrefix, func(key kvstore.Key, value kvstore.Value) bool {
		ret[Key(key[prefix:])] = value
		return true
	})
	if err != nil {
		panic(asDBError(err))
	}
	b.mutations.Iterate(func(key Key, mut Mutation) bool {
		v := mut.Value()
		if v != nil {
			ret[key] = v
		} else {
			delete(ret, key)
		}
		return true
	})
	return ret
}

// iterates over all key-value pairs in KVStore
func (b *bufferedKVStore) DangerouslyDumpToString() string {
	ret := "         BufferedKVStore:\n"
	for k, v := range b.DangerouslyDumpToMap() {
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
	mut := b.mutations.Get(k)
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
	mut := b.mutations.Get(key)
	if mut != nil {
		return mut.Value(), nil
	}
	v, err := b.db.Get(kvstore.Key(key))
	if err == kvstore.ErrKeyNotFound {
		return nil, nil
	}
	return v, asDBError(err)
}
