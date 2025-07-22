// Package dict provides a dictionary implementation for key-value storage.
// It implements the kv.KVStore interface with an in-memory map as the backend,
// allowing for efficient key-value operations and serialization.
package dict

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"sort"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/hashing"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/util/rwutil"
)

// Dict is an implementation kv.KVStore interface backed by an in-memory map.
// kv.KVStore represents a key-value store
// where both keys and values are arbitrary byte slices.
type Dict map[kv.Key][]byte

// New creates new
func New() Dict {
	return make(Dict)
}

// Clone creates clone (deep copy) of Dict
func (d Dict) Clone() Dict {
	clone := make(Dict)
	d.ForEach(func(key kv.Key, value []byte) bool {
		clone.Set(key, slices.Clone(value))
		return true
	})
	return clone
}

// FromKVStore convert (copy) any KVStore to dict
func FromKVStore(s kv.KVStoreReader) Dict {
	d := make(Dict)
	s.Iterate(kv.EmptyPrefix, func(k kv.Key, v []byte) bool {
		d[k] = v
		return true
	})
	return d
}

func (d Dict) String() string {
	ret := "         Dict:\n"
	for _, key := range d.KeysSorted() {
		val := d[key]
		if len(val) > 80 {
			val = val[:80]
		}
		ret += fmt.Sprintf("           %q: %q\n", key, val)
	}
	return ret
}

// ForEach iterates non-deterministic!
func (d Dict) ForEach(fun func(key kv.Key, value []byte) bool) {
	for k, v := range d {
		if !fun(k, v) {
			return // abort when callback returns false
		}
	}
}

// IsEmpty returns of it has no records
func (d Dict) IsEmpty() bool {
	return len(d) == 0
}

// Set sets the value for the key
func (d Dict) Set(key kv.Key, value []byte) {
	if value == nil {
		panic("cannot Set(key, nil), use Del() to remove a key/value")
	}
	d[key] = value
}

// Del removes key/value pair
func (d Dict) Del(key kv.Key) {
	delete(d, key)
}

// Has checks if key exist
func (d Dict) Has(key kv.Key) bool {
	_, ok := d[key]
	return ok
}

// Iterate over keys with prefix
func (d Dict) Iterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	d.IterateKeys(prefix, func(key kv.Key) bool {
		return f(key, d[key])
	})
}

// IterateKeys over keys with prefix
func (d Dict) IterateKeys(prefix kv.Key, f func(key kv.Key) bool) {
	for k := range d {
		if !k.HasPrefix(prefix) {
			continue
		}
		if !f(k) {
			break
		}
	}
}

func (d Dict) IterateSorted(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	d.IterateKeysSorted(prefix, func(key kv.Key) bool {
		return f(key, d[key])
	})
}

func (d Dict) IterateKeysSorted(prefix kv.Key, f func(key kv.Key) bool) {
	for _, k := range d.KeysSorted() {
		if !k.HasPrefix(prefix) {
			continue
		}
		if !f(k) {
			break
		}
	}
}

// Get takes a value. Returns nil if key does not exist
func (d Dict) Get(key kv.Key) []byte {
	return d[key]
}

func (d Dict) Bytes() []byte {
	return rwutil.WriteToBytes(&d)
}

func FromBytes(data []byte) (ret Dict, err error) {
	ret = New()
	_, err = rwutil.ReadFromBytes(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret, err
}

func (d *Dict) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	size := rr.ReadSize32()
	for i := 0; i < size; i++ {
		key := kv.Key(rr.ReadBytes())
		value := rr.ReadBytes()
		if rr.Err == nil {
			d.Set(key, value)
		}
	}
	return rr.Err
}

func (d *Dict) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	keys := d.KeysSorted()
	ww.WriteSize32(len(keys))
	for _, key := range keys {
		ww.WriteBytes([]byte(key))
		ww.WriteBytes(d.Get(key))
	}
	return ww.Err
}

// Keys takes all keys
func (d Dict) Keys() []kv.Key {
	ret := make([]kv.Key, 0)
	for key := range d {
		ret = append(ret, key)
	}
	return ret
}

// KeysSorted takes keys and sorts them
func (d Dict) KeysSorted() []kv.Key {
	k := d.Keys()
	sort.Slice(k, func(i, j int) bool {
		return k[i] < k[j]
	})
	return k
}

// Extend appends another Dict
func (d Dict) Extend(from Dict) {
	for key, value := range from {
		d.Set(key, value)
	}
}

// Hash takes deterministic has of the dict
func (d Dict) Hash() hashing.HashValue {
	keys := d.KeysSorted()
	data := make([][]byte, 0, 2*len(d))
	for _, k := range keys {
		data = append(data, []byte(k))
		data = append(data, d.Get(k))
	}
	return hashing.HashData(data...)
}

func (d Dict) Equals(d1 Dict) bool {
	if len(d) != len(d1) {
		return false
	}
	for k, v := range d {
		v1, ok := d1[k]
		if !ok {
			return false
		}
		if !bytes.Equal(v, v1) {
			return false
		}
	}
	return true
}

// JSONDict is the JSON-compatible representation of a Dict
type jsonDict struct {
	Items []item
}

// Item is a JSON-compatible representation of a single key-value pair
type item struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (d Dict) MarshalJSON() ([]byte, error) {
	j := jsonDict{Items: make([]item, len(d))}

	for i, k := range d.KeysSorted() {
		j.Items[i].Key = hexutil.Encode([]byte(k))
		j.Items[i].Value = hexutil.Encode(d[k])
	}

	return json.Marshal(j)
}

func (d *Dict) UnmarshalJSON(b []byte) error {
	var j jsonDict
	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}
	*d = make(Dict)
	for _, item := range j.Items {
		k, err := cryptolib.DecodeHex(item.Key)
		if err != nil {
			return err
		}
		v, err := cryptolib.DecodeHex(item.Value)
		if err != nil {
			return err
		}
		(*d)[kv.Key(k)] = v
	}
	return nil
}
