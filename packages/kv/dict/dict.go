package dict

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/iotaledger/wasp/packages/hashing"
	"io"
	"sort"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/mr-tron/base58"
)

// Dict is a KVStore backed by an in-memory map
type Dict map[kv.Key][]byte

func (d Dict) MustGet(key kv.Key) []byte {
	return kv.MustGet(d, key)
}

func (d Dict) MustHas(key kv.Key) bool {
	return kv.MustHas(d, key)
}

func (d Dict) MustIterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	kv.MustIterate(d, prefix, f)
}

func (d Dict) MustIterateKeys(prefix kv.Key, f func(key kv.Key) bool) {
	kv.MustIterateKeys(d, prefix, f)
}

// create/clone
func New() Dict {
	return make(Dict)
}

func (d Dict) Clone() Dict {
	clone := make(Dict)
	d.ForEach(func(key kv.Key, value []byte) bool {
		clone.Set(key, value)
		return true
	})
	return clone
}

func FromGoMap(d map[kv.Key][]byte) Dict {
	return Dict(d)
}

func Clone(other Dict) (Dict, error) {
	d := make(Dict)
	for k, v := range other {
		d[k] = v
	}
	return d, nil
}

func FromKVStore(s kv.KVStore) (Dict, error) {
	d := make(Dict)
	err := s.Iterate(kv.EmptyPrefix, func(k kv.Key, v []byte) bool {
		d[k] = v
		return true
	})
	return d, err
}

func (d Dict) sortedKeys() []kv.Key {
	keys := make([]kv.Key, 0)
	for k := range d {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return keys
}

func (d Dict) String() string {
	ret := "         Dict:\n"
	for _, key := range d.sortedKeys() {
		val := d[key]
		if len(val) > 80 {
			val = val[:80]
		}
		ret += fmt.Sprintf(
			"           0x%s: 0x%s (base58: %s) ('%s': '%s')\n",
			slice(hex.EncodeToString([]byte(key))),
			slice(hex.EncodeToString(val)),
			slice(base58.Encode(val)),
			string(key),
			string(val),
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

// NON DETERMINISTIC!
func (d Dict) ForEach(fun func(key kv.Key, value []byte) bool) {
	for k, v := range d {
		if !fun(k, v) {
			return // abort when callback returns false
		}
	}
}

func (d Dict) ForEachDeterministic(fun func(key kv.Key, value []byte) bool) {
	if d == nil {
		return
	}
	for _, k := range d.sortedKeys() {
		if !fun(k, d[k]) {
			return // abort when callback returns false
		}
	}
}

func (d Dict) IsEmpty() bool {
	return len(d) == 0
}

func (d Dict) Set(key kv.Key, value []byte) {
	if value == nil {
		panic("cannot Set(key, nil), use Del() to remove a key/value")
	}
	d[key] = value
}

func (d Dict) Del(key kv.Key) {
	delete(d, key)
}

func (d Dict) Has(key kv.Key) (bool, error) {
	_, ok := d[key]
	return ok, nil
}

func (d Dict) Iterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) error {
	for k, v := range d {
		if !k.HasPrefix(prefix) {
			continue
		}
		if !f(k, v) {
			break
		}
	}
	return nil
}

func (d Dict) IterateKeys(prefix kv.Key, f func(key kv.Key) bool) error {
	for k := range d {
		if !k.HasPrefix(prefix) {
			continue
		}
		if !f(k) {
			break
		}
	}
	return nil
}

func (d Dict) Get(key kv.Key) ([]byte, error) {
	v, _ := d[key]
	return v, nil
}

func (d Dict) Write(w io.Writer) error {
	keys := d.sortedKeys()
	if err := util.WriteUint64(w, uint64(len(keys))); err != nil {
		return err
	}
	for _, k := range keys {
		if err := util.WriteBytes16(w, []byte(k)); err != nil {
			return err
		}
		if err := util.WriteBytes32(w, d[k]); err != nil {
			return err
		}
	}
	return nil
}

func (d Dict) Read(r io.Reader) error {
	var num uint64
	err := util.ReadUint64(r, &num)
	if err != nil {
		return err
	}
	for i := uint64(0); i < num; i++ {
		k, err := util.ReadBytes16(r)
		if err != nil {
			return err
		}
		v, err := util.ReadBytes32(r)
		if err != nil {
			return err
		}
		d.Set(kv.Key(k), v)
	}
	return nil
}

func (d Dict) Keys() []kv.Key {
	ret := make([]kv.Key, 0)
	for key := range d {
		ret = append(ret, key)
	}
	return ret
}

func (d Dict) KeysSorted() []kv.Key {
	k := d.Keys()
	sort.Slice(k, func(i, j int) bool {
		return k[i] < k[j]
	})
	return k
}

func (d Dict) Extend(from Dict) {
	for key, value := range from {
		d.Set(key, value)
	}
}

func (d Dict) Hash() hashing.HashValue {
	keys := d.KeysSorted()
	data := make([][]byte, 0, 2*len(d))
	for _, k := range keys {
		data = append(data, []byte(k))
		v, _ := d.Get(k)
		data = append(data, v)
	}
	return *hashing.HashData(data...)
}

type jsonItem struct {
	Key   []byte
	Value []byte
}

func (d Dict) MarshalJSON() ([]byte, error) {
	items := make([]jsonItem, len(d))
	for i, k := range d.sortedKeys() {
		items[i].Key = []byte(k)
		items[i].Value = d[k]
	}
	return json.Marshal(items)
}

func (d *Dict) UnmarshalJSON(b []byte) error {
	var items []jsonItem
	if err := json.Unmarshal(b, &items); err != nil {
		return err
	}
	*d = make(Dict)
	for _, item := range items {
		(*d)[kv.Key(item.Key)] = item.Value
	}
	return nil
}
