package buffered

import (
	"fmt"
	"sort"

	iotago "github.com/iotaledger/iota.go/v3"
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

// DangerouslyDumpToDict returns a Dict with the whole contents of the
// backing store + applied mutations.
func (b *BufferedKVStore) DangerouslyDumpToDict() dict.Dict {
	ret := dict.New()
	err := b.Iterate("", func(key kv.Key, value []byte) bool {
		ret.Set(key, value)
		return true
	})
	if err != nil {
		panic(err)
	}
	return ret
}

// iterates over all key-value pairs in KVStore
func (b *BufferedKVStore) DangerouslyDumpToString() string {
	ret := "         BufferedKVStore:\n"
	for k, v := range b.DangerouslyDumpToDict() {
		ret += fmt.Sprintf(
			"           [%s] %s: %s\n",
			b.flag(k),
			slice(iotago.EncodeHex([]byte(k))),
			slice(iotago.EncodeHex(v)),
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

func (b *BufferedKVStore) Get(key kv.Key) ([]byte, error) {
	v, ok := b.muts.Get(key)
	if ok {
		return v, nil
	}
	return b.r.Get(key)
}

func (b *BufferedKVStore) MustGet(key kv.Key) []byte {
	return kv.MustGet(b, key)
}

func (b *BufferedKVStore) Has(key kv.Key) (bool, error) {
	v, ok := b.muts.Get(key)
	if ok {
		return v != nil, nil
	}
	return b.r.Has(key)
}

func (b *BufferedKVStore) MustHas(key kv.Key) bool {
	return kv.MustHas(b, key)
}

func (b *BufferedKVStore) Iterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) error {
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

func (b *BufferedKVStore) MustIterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	kv.MustIterate(b, prefix, f)
}

func (b *BufferedKVStore) IterateKeys(prefix kv.Key, f func(key kv.Key) bool) error {
	for k := range b.muts.Sets {
		if !k.HasPrefix(prefix) {
			continue
		}
		if !f(k) {
			return nil
		}
	}
	return b.r.IterateKeys(prefix, func(k kv.Key) bool {
		if !b.muts.Contains(k) {
			return f(k)
		}
		return true
	})
}

func (b *BufferedKVStore) IterateSorted(prefix kv.Key, f func(key kv.Key, value []byte) bool) error {
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

func (b *BufferedKVStore) MustIterateSorted(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	kv.MustIterateSorted(b, prefix, f)
}

func (b *BufferedKVStore) IterateKeysSorted(prefix kv.Key, f func(key kv.Key) bool) error {
	var keys []kv.Key

	for k := range b.muts.Sets {
		if !k.HasPrefix(prefix) {
			continue
		}
		keys = append(keys, k)
	}

	err := b.r.IterateKeysSorted(prefix, func(k kv.Key) bool {
		if !b.muts.Contains(k) {
			keys = append(keys, k)
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

func (b *BufferedKVStore) MustIterateKeys(prefix kv.Key, f func(key kv.Key) bool) {
	kv.MustIterateKeys(b, prefix, f)
}

func (b *BufferedKVStore) MustIterateKeysSorted(prefix kv.Key, f func(key kv.Key) bool) {
	kv.MustIterateKeysSorted(b, prefix, f)
}
