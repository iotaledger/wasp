package dict

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/mr-tron/base58"
)

// Dict is a KVStore backed by an in-memory map
type Dict interface {
	kv.KVStore

	IsEmpty() bool
	Len() int

	ForEach(func(key kv.Key, value []byte) bool)
	ForEachDeterministic(func(key kv.Key, value []byte) bool)
	Clone() Dict

	Read(io.Reader) error
	Write(io.Writer) error

	String() string
}

type dict map[kv.Key][]byte

// create/clone
func New() Dict {
	return make(dict)
}

func (d dict) Clone() Dict {
	clone := make(dict)
	d.ForEach(func(key kv.Key, value []byte) bool {
		clone.Set(key, value)
		return true
	})
	return clone
}

func FromGoMap(d map[kv.Key][]byte) Dict {
	return dict(d)
}

func FromKVStore(kvstore kv.KVStore) (Dict, error) {
	d := make(dict)
	err := kvstore.Iterate(kv.EmptyPrefix, func(k kv.Key, v []byte) bool {
		d[k] = v
		return true
	})
	return d, err
}

func (d dict) sortedKeys() []kv.Key {
	keys := make([]kv.Key, 0)
	for k := range d {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return keys
}

func (d dict) String() string {
	ret := "         Dict:\n"
	for _, key := range d.sortedKeys() {
		ret += fmt.Sprintf(
			"           0x%s: 0x%s (base58: %s) ('%s': '%s')\n",
			slice(hex.EncodeToString([]byte(key))),
			slice(hex.EncodeToString(d[key])),
			slice(base58.Encode(d[key])),
			string(key),
			string(d[key]),
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
func (d dict) ForEach(fun func(key kv.Key, value []byte) bool) {
	for k, v := range d {
		if !fun(k, v) {
			return // abort when callback returns false
		}
	}
}

func (d dict) ForEachDeterministic(fun func(key kv.Key, value []byte) bool) {
	if d == nil {
		return
	}
	for _, k := range d.sortedKeys() {
		if !fun(k, d[k]) {
			return // abort when callback returns false
		}
	}
}

func (d dict) IsEmpty() bool {
	return len(d) == 0
}

func (d dict) Len() int {
	return len(d)
}

func (d dict) Set(key kv.Key, value []byte) {
	if value == nil {
		panic("cannot Set(key, nil), use Del() to remove a key/value")
	}
	d[key] = value
}

func (d dict) Del(key kv.Key) {
	delete(d, key)
}

func (d dict) Has(key kv.Key) (bool, error) {
	_, ok := d[key]
	return ok, nil
}

func (d dict) Iterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) error {
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

func (d dict) IterateKeys(prefix kv.Key, f func(key kv.Key) bool) error {
	for k, _ := range d {
		if !k.HasPrefix(prefix) {
			continue
		}
		if !f(k) {
			break
		}
	}
	return nil
}

func (d dict) Get(key kv.Key) ([]byte, error) {
	v, _ := d[key]
	return v, nil
}

func (d dict) Write(w io.Writer) error {
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

func (d dict) Read(r io.Reader) error {
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

type jsonItem struct {
	Key   []byte
	Value []byte
}

func (d dict) MarshalJSON() ([]byte, error) {
	items := make([]jsonItem, d.Len())
	for i, k := range d.sortedKeys() {
		items[i].Key = []byte(k)
		items[i].Value = d[k]
	}
	return json.Marshal(items)
}

func (d *dict) UnmarshalJSON(b []byte) error {
	var items []jsonItem
	if err := json.Unmarshal(b, &items); err != nil {
		return err
	}
	for _, item := range items {
		(*d)[kv.Key(item.Key)] = item.Value
	}
	return nil
}
