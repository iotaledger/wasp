// Package test provides testing utilities for the trie package.
package test

import (
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/kvstore"
	"github.com/iotaledger/wasp/v2/packages/kvstore/mapdb"
	"github.com/iotaledger/wasp/v2/packages/trie"
	"github.com/iotaledger/wasp/v2/packages/util"
)

// ----------------------------------------------------------------------------
// InMemoryKVStore is a KVStore implementation. Mostly used for testing
var (
	_ trie.KVStore = &InMemoryKVStore{}
)

type (
	InMemoryKVStore struct {
		m     kvstore.KVStore
		Stats InMemoryKVStoreStats
	}

	InMemoryKVStoreStats struct {
		Get          uint
		MultiGet     uint
		MultiGetKeys uint
		Has          uint
		Set          uint
		Del          uint
	}
)

func NewInMemoryKVStore() *InMemoryKVStore {
	return &InMemoryKVStore{
		m: mapdb.NewMapDB(),
	}
}

func (im *InMemoryKVStore) ResetStats() {
	im.Stats = InMemoryKVStoreStats{}
}

func (im *InMemoryKVStore) Get(k []byte) []byte {
	im.Stats.Get++
	v, err := im.m.Get(k)
	if err == kvstore.ErrKeyNotFound {
		return nil
	}
	lo.Must0(err)
	return v
}

func (im *InMemoryKVStore) MultiGet(ks [][]byte) [][]byte {
	im.Stats.MultiGet++
	im.Stats.MultiGetKeys += uint(len(ks))
	return lo.Must(im.m.MultiGet(ks))
}

func (im *InMemoryKVStore) Has(k []byte) bool {
	im.Stats.Has++
	return lo.Must(im.m.Has(k))
}

func (im *InMemoryKVStore) Iterate(prefix []byte, f func(k []byte, v []byte) bool) {
	lo.Must0(im.m.Iterate(prefix, f))
}

func (im *InMemoryKVStore) IterateKeys(prefix []byte, f func(k []byte) bool) {
	lo.Must0(im.m.IterateKeys(prefix, f))
}

func (im *InMemoryKVStore) Set(k, v []byte) {
	im.Stats.Set++
	if len(v) != 0 {
		lo.Must0(im.m.Set(k, v))
	} else {
		lo.Must0(im.m.Delete(k))
	}
}

func (im *InMemoryKVStore) Del(k []byte) {
	im.Stats.Del++
	lo.Must0(im.m.Delete(k))
}

// RandStreamIterator is a stream of random key/value pairs with the given parameters
// Used for testing
var _ kv.StreamIterator = &PseudoRandStreamIterator{}

type PseudoRandStreamIterator struct {
	rnd   *rand.Rand
	par   PseudoRandStreamParams
	count int
}

// PseudoRandStreamParams represents parameters of the RandStreamIterator
type PseudoRandStreamParams struct {
	// Seed for deterministic randomization
	Seed int64
	// NumKVPairs maximum number of key value pairs to generate. 0 means infinite
	NumKVPairs int
	// MaxKey maximum length of key (randomly generated)
	MaxKey int
	// MaxValue maximum length of value (randomly generated)
	MaxValue int
}

func NewPseudoRandStreamIterator(p ...PseudoRandStreamParams) *PseudoRandStreamIterator {
	ret := &PseudoRandStreamIterator{
		par: PseudoRandStreamParams{
			Seed:       time.Now().UnixNano() + int64(os.Getpid()),
			NumKVPairs: 0, // infinite
			MaxKey:     64,
			MaxValue:   128,
		},
	}
	if len(p) > 0 {
		ret.par = p[0]
	}
	ret.rnd = util.NewPseudoRand(ret.par.Seed)
	return ret
}

func (r *PseudoRandStreamIterator) Iterate(fun func(k []byte, v []byte) bool) error {
	maxNumKVPairs := r.par.NumKVPairs
	if maxNumKVPairs <= 0 {
		maxNumKVPairs = math.MaxInt
	}
	for r.count < maxNumKVPairs {
		k := make([]byte, r.rnd.Intn(r.par.MaxKey-1)+1)
		r.rnd.Read(k)
		v := make([]byte, r.rnd.Intn(r.par.MaxValue-1)+1)
		r.rnd.Read(v)
		if !fun(k, v) {
			return nil
		}
		r.count++
	}
	return nil
}
