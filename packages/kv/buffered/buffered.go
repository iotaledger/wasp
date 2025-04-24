package buffered

import (
	"fmt"
	"sort"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// BufferedKVStore is a KVStore backed by a given KVStoreReader. Writes are cached in-memory;
// reads are delegated to the backing store for unmodified keys.
type BufferedKVStore struct {
	r    kv.KVStoreReader
	muts *Mutations
}

func NewBufferedKVStore(r kv.KVStoreReader) *BufferedKVStore {
	return &BufferedKVStore{
		r:    r,
		muts: NewMutations(),
	}
}

func NewBufferedKVStoreForMutations(r kv.KVStoreReader, m *Mutations) *BufferedKVStore {
	return &BufferedKVStore{
		r:    r,
		muts: m,
	}
}

func (b *BufferedKVStore) Clone() *BufferedKVStore {
	return &BufferedKVStore{
		r:    b.r,
		muts: b.muts.Clone(),
	}
}

func (b *BufferedKVStore) Mutations() *Mutations {
	return b.muts
}

func (b *BufferedKVStore) MutationsCount() int {
	return len(b.muts.Sets) + len(b.muts.Dels)
}

func (b *BufferedKVStore) SetMutations(muts *Mutations) {
	b.muts = muts
}

func (b *BufferedKVStore) SetKVStoreReader(r kv.KVStoreReader) {
	b.r = r
}

// DangerouslyDumpToDict returns a Dict with the whole contents of the
// backing store + applied mutations.
func (b *BufferedKVStore) DangerouslyDumpToDict() dict.Dict {
	ret := dict.New()
	b.Iterate("", func(key kv.Key, value []byte) bool {
		ret.Set(key, value)
		return true
	})
	return ret
}

// iterates over all key-value pairs in KVStore
func (b *BufferedKVStore) DangerouslyDumpToString() string {
	ret := "         BufferedKVStore:\n"
	for k, v := range b.DangerouslyDumpToDict() {
		ret += fmt.Sprintf(
			"           [%s] %s: %s\n",
			b.flag(k),
			slice(hexutil.Encode([]byte(k))),
			slice(hexutil.Encode(v)),
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

func (b *BufferedKVStore) flag(k kv.Key) string {
	if b.muts.Contains(k) {
		return "+"
	}
	return " "
}

func (b *BufferedKVStore) Set(key kv.Key, value []byte) {
	b.muts.Set(key, value)
}

func (b *BufferedKVStore) Del(key kv.Key) {
	b.muts.Del(key)
}

func (b *BufferedKVStore) Get(key kv.Key) []byte {
	v, ok := b.muts.Get(key)
	if ok {
		return v
	}
	return b.r.Get(key)
}

func (b *BufferedKVStore) Has(key kv.Key) bool {
	v, ok := b.muts.Get(key)
	if ok {
		return v != nil
	}
	return b.r.Has(key)
}

func (b *BufferedKVStore) Iterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	b.IterateKeys(prefix, func(k kv.Key) bool {
		return f(k, b.Get(k))
	})
}

func (b *BufferedKVStore) IterateKeys(prefix kv.Key, f func(key kv.Key) bool) {
	for k := range b.muts.Sets {
		if !k.HasPrefix(prefix) {
			continue
		}
		if !f(k) {
			return
		}
	}
	b.r.IterateKeys(prefix, func(k kv.Key) bool {
		if !b.muts.Contains(k) {
			return f(k)
		}
		return true
	})
}

func (b *BufferedKVStore) IterateSorted(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	b.IterateKeysSorted(prefix, func(k kv.Key) bool {
		return f(k, b.Get(k))
	})
}

func (b *BufferedKVStore) IterateKeysSorted(prefix kv.Key, f func(key kv.Key) bool) {
	var keys []kv.Key

	for k := range b.muts.Sets {
		if !k.HasPrefix(prefix) {
			continue
		}
		keys = append(keys, k)
	}

	b.r.IterateKeysSorted(prefix, func(k kv.Key) bool {
		if !b.muts.Contains(k) {
			keys = append(keys, k)
		}
		return true
	})

	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	for _, k := range keys {
		if !f(k) {
			break
		}
	}
}
