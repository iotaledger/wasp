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

type DB interface {
	Get(key Key) ([]byte, error)
	Iterate(func(key Key, value []byte) bool) error
}

type dbtable struct {
	db        DB
	mutations MutationMap
}

func NewBufferedKVStore(db DB) BufferedKVStore {
	return &dbtable{
		db:        db,
		mutations: NewMutationMap(),
	}
}

func NewBufferedKVStoreOnSubrealm(db func() kvstore.KVStore, prefix kvstore.KeyPrefix) BufferedKVStore {
	return NewBufferedKVStore(&subrealm{
		db:     db,
		prefix: prefix,
	})
}

func (c *dbtable) Clone() BufferedKVStore {
	return &dbtable{
		db:        c.db,
		mutations: c.mutations.Clone(),
	}
}

func (c *dbtable) Codec() Codec {
	return NewCodec(c)
}

func (c *dbtable) Mutations() MutationMap {
	return c.mutations
}

func (c *dbtable) ClearMutations() {
	c.mutations = NewMutationMap()
}

// iterates over all key-value pairs in KVStore
func (c *dbtable) DangerouslyDumpToMap() map[Key][]byte {
	ret := make(map[Key][]byte)
	err := c.db.Iterate(func(key Key, value []byte) bool {
		ret[Key(key)] = value
		return true
	})
	if err != nil {
		panic(err)
	}
	c.mutations.Iterate(func(key Key, mut Mutation) bool {
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
func (c *dbtable) DangerouslyDumpToString() string {
	ret := "         BufferedKVStore:\n"
	for k, v := range c.DangerouslyDumpToMap() {
		ret += fmt.Sprintf(
			"           [%s] 0x%s: 0x%s (base58: %s)\n",
			c.flag(k),
			slice(hex.EncodeToString([]byte(k))),
			slice(hex.EncodeToString(v)),
			slice(base58.Encode(v)),
		)
	}
	return ret
}

func (c *dbtable) flag(k Key) string {
	mut := c.mutations.Get(k)
	if mut != nil {
		return "+"
	}
	return " "
}

func (c *dbtable) Set(key Key, value []byte) {
	c.mutations.Add(NewMutationSet(key, value))
}

func (c *dbtable) Del(key Key) {
	c.mutations.Add(NewMutationDel(key))
}

func (c *dbtable) Get(key Key) ([]byte, error) {
	mut := c.mutations.Get(key)
	if mut != nil {
		return mut.Value(), nil
	}
	b, err := c.db.Get(key)
	return b, err
}

type subrealm struct {
	db     func() kvstore.KVStore
	prefix kvstore.KeyPrefix
}

func (s *subrealm) Get(key Key) ([]byte, error) {
	v, err := s.db().Get(append(s.prefix, []byte(key)...))
	if err == kvstore.ErrKeyNotFound {
		return nil, nil
	}
	return v, err
}

func (s *subrealm) Iterate(f func(Key, []byte) bool) error {
	db := s.db()
	return db.Iterate(s.prefix, func(key kvstore.Key, value kvstore.Value) bool {
		return f(Key(key[len(db.Realm())+len(s.prefix):]), value)
	})
}
