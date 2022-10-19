package buffered

import (
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// BufferedKVStoreAccess is a KVStore backed by a given KVStoreReader. Writes are cached in-memory;
// reads are delegated to the backing store for unmodified keys.
type BufferedKVStoreAccess struct {
	r    kv.KVStoreReader
	muts *Mutations
}

func NewBufferedKVStoreAccess(r kv.KVStoreReader) *BufferedKVStoreAccess {
	return &BufferedKVStoreAccess{
		r:    r,
		muts: NewMutations(),
	}
}

func (b *BufferedKVStoreAccess) Copy() *BufferedKVStoreAccess {
	return &BufferedKVStoreAccess{
		r:    b.r,
		muts: b.muts.Clone(),
	}
}

func (b *BufferedKVStoreAccess) Mutations() *Mutations {
	return b.muts
}

func (b *BufferedKVStoreAccess) ClearMutations() {
	b.muts = NewMutations()
}

// DangerouslyDumpToDict returns a Dict with the whole contents of the
// backing store + applied mutations.
func (b *BufferedKVStoreAccess) DangerouslyDumpToDict() dict.Dict {
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
func (b *BufferedKVStoreAccess) DangerouslyDumpToString() string {
	ret := "         BufferedKVStoreAccess:\n"
	for k, v := range b.DangerouslyDumpToDict() {
		ret += fmt.Sprintf(
			"           [%s] 0x%s: 0x%s (hex: %s)\n",
			b.flag(k),
			slice(hex.EncodeToString([]byte(k))),
			slice(hex.EncodeToString(v)),
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

func (b *BufferedKVStoreAccess) flag(k kv.Key) string {
	if b.muts.Contains(k) {
		return "+"
	}
	return " "
}

func (b *BufferedKVStoreAccess) Set(key kv.Key, value []byte) {
	b.muts.Set(key, value)
}

func (b *BufferedKVStoreAccess) Del(key kv.Key) {
	b.muts.Del(key)
}

func (b *BufferedKVStoreAccess) Get(key kv.Key) ([]byte, error) {
	v, ok := b.muts.Get(key)
	if ok {
		return v, nil
	}
	return b.r.Get(key)
}

func (b *BufferedKVStoreAccess) MustGet(key kv.Key) []byte {
	return kv.MustGet(b, key)
}

func (b *BufferedKVStoreAccess) Has(key kv.Key) (bool, error) {
	v, ok := b.muts.Get(key)
	if ok {
		return v != nil, nil
	}
	return b.r.Has(key)
}

func (b *BufferedKVStoreAccess) MustHas(key kv.Key) bool {
	return kv.MustHas(b, key)
}

func (b *BufferedKVStoreAccess) Iterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) error {
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

func (b *BufferedKVStoreAccess) MustIterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	kv.MustIterate(b, prefix, f)
}

func (b *BufferedKVStoreAccess) IterateKeys(prefix kv.Key, f func(key kv.Key) bool) error {
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

func (b *BufferedKVStoreAccess) IterateSorted(prefix kv.Key, f func(key kv.Key, value []byte) bool) error {
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

func (b *BufferedKVStoreAccess) MustIterateSorted(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	kv.MustIterateSorted(b, prefix, f)
}

func (b *BufferedKVStoreAccess) IterateKeysSorted(prefix kv.Key, f func(key kv.Key) bool) error {
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

func (b *BufferedKVStoreAccess) MustIterateKeys(prefix kv.Key, f func(key kv.Key) bool) {
	kv.MustIterateKeys(b, prefix, f)
}

func (b *BufferedKVStoreAccess) MustIterateKeysSorted(prefix kv.Key, f func(key kv.Key) bool) {
	kv.MustIterateKeysSorted(b, prefix, f)
}
