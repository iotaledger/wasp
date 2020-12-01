package buffered

import (
	"encoding/hex"
	"fmt"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/mr-tron/base58"
)

// BufferedKVStore represents a KVStore backed by a database. Writes are cached in-memory as
// a MutationSequence; reads are delegated to the backing database when not cached.
type BufferedKVStore interface {
	kv.KVStore

	// the uncommitted mutations
	Mutations() MutationSequence
	ClearMutations()
	Clone() BufferedKVStore

	Codec() codec.MutableCodec

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

func (b *bufferedKVStore) Codec() codec.MutableCodec {
	return codec.NewCodec(b)
}

func (b *bufferedKVStore) Mutations() MutationSequence {
	return b.mutations
}

func (b *bufferedKVStore) ClearMutations() {
	b.mutations = NewMutationSequence()
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
	for k, v := range kv.ToGoMap(b.DangerouslyDumpToDict()) {
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
	mut := b.mutations.Latest(k)
	if mut != nil {
		return "+"
	}
	return " "
}

func (b *bufferedKVStore) Set(key kv.Key, value []byte) {
	b.mutations.Add(NewMutationSet(key, value))
}

func (b *bufferedKVStore) Del(key kv.Key) {
	b.mutations.Add(NewMutationDel(key))
}

func (b *bufferedKVStore) Get(key kv.Key) ([]byte, error) {
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

func (b *bufferedKVStore) Has(key kv.Key) (bool, error) {
	mut := b.mutations.Latest(key)
	if mut != nil {
		return mut.Value() != nil, nil
	}
	v, err := b.db.Has(kvstore.Key(key))
	return v, asDBError(err)
}

func (b *bufferedKVStore) Iterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) error {
	seen, done := b.mutations.IterateValues(prefix, f)
	if done {
		return nil
	}
	return b.db.Iterate([]byte(prefix), func(key kvstore.Key, value kvstore.Value) bool {
		k := kv.Key(key)
		_, ok := seen[k]
		if ok {
			return true
		}
		return f(k, value)
	})
}

func (b *bufferedKVStore) IterateKeys(prefix kv.Key, f func(key kv.Key) bool) error {
	seen, done := b.mutations.IterateValues(prefix, func(key kv.Key, value []byte) bool {
		return f(key)
	})
	if done {
		return nil
	}
	return b.db.IterateKeys([]byte(prefix), func(key kvstore.Key) bool {
		k := kv.Key(key)
		_, ok := seen[k]
		if ok {
			return true
		}
		return f(k)
	})
}
