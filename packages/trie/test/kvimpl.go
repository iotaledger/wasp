// Package test provides testing utilities for the trie package.
package test

import (
	"bytes"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/trie"
	"github.com/iotaledger/wasp/v2/packages/util"
)

// ----------------------------------------------------------------------------
// InMemoryKVStore is a KVStore implementation. Mostly used for testing
var (
	_ trie.KVStore     = &InMemoryKVStore{}
	_ trie.Traversable = &InMemoryKVStore{}
	_ trie.KVIterator  = &simpleInMemoryIterator{}
)

type (
	InMemoryKVStore struct {
		m     map[string][]byte
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

	simpleInMemoryIterator struct {
		store  *InMemoryKVStore
		prefix []byte
	}
)

func NewInMemoryKVStore() *InMemoryKVStore {
	return &InMemoryKVStore{
		m: make(map[string][]byte),
	}
}

func (im *InMemoryKVStore) ResetStats() {
	im.Stats = InMemoryKVStoreStats{}
}

func (im *InMemoryKVStore) Get(k []byte) []byte {
	im.Stats.Get++
	return im.m[string(k)]
}

func (im *InMemoryKVStore) MultiGet(ks [][]byte) [][]byte {
	im.Stats.MultiGet++
	im.Stats.MultiGetKeys += uint(len(ks))
	return lo.Map(ks, func(k []byte, _ int) []byte {
		return im.m[string(k)]
	})
}

func (im *InMemoryKVStore) Has(k []byte) bool {
	im.Stats.Has++
	_, ok := im.m[string(k)]
	return ok
}

func (im *InMemoryKVStore) Iterate(f func(k []byte, v []byte) bool) {
	for k, v := range im.m {
		if !f([]byte(k), v) {
			return
		}
	}
}

func (im *InMemoryKVStore) IterateKeys(f func(k []byte) bool) {
	for k := range im.m {
		if !f([]byte(k)) {
			return
		}
	}
}

func (im *InMemoryKVStore) Set(k, v []byte) {
	im.Stats.Set++
	if len(v) != 0 {
		im.m[string(k)] = v
	} else {
		delete(im.m, string(k))
	}
}

func (im *InMemoryKVStore) Del(k []byte) {
	im.Stats.Del++
	delete(im.m, string(k))
}

func (im *InMemoryKVStore) Iterator(prefix []byte) trie.KVIterator {
	return &simpleInMemoryIterator{
		store:  im,
		prefix: prefix,
	}
}

func (si *simpleInMemoryIterator) Iterate(f func(k []byte, v []byte) bool) {
	var key []byte
	for k, v := range si.store.m {
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
	for k := range si.store.m {
		key = []byte(k)
		if bytes.HasPrefix(key, si.prefix) {
			if !f(key) {
				return
			}
		}
	}
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
