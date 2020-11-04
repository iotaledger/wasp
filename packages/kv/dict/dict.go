package dict

import (
	"encoding/hex"
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

	ForEach(func(key kv.Key, value []byte) bool)
	ForEachDeterministic(func(key kv.Key, value []byte) bool)
	Clone() Dict

	Read(io.Reader) error
	Write(io.Writer) error

	String() string
}

type kvmap map[kv.Key][]byte

// create/clone
func NewDict() Dict {
	return make(kvmap)
}

func (m kvmap) Clone() Dict {
	clone := make(kvmap)
	m.ForEach(func(key kv.Key, value []byte) bool {
		clone.Set(key, value)
		return true
	})
	return clone
}

func FromGoMap(m map[kv.Key][]byte) Dict {
	return kvmap(m)
}

func (m kvmap) sortedKeys() []kv.Key {
	keys := make([]kv.Key, 0)
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return keys
}

func (m kvmap) String() string {
	ret := "         Dict:\n"
	for _, key := range m.sortedKeys() {
		ret += fmt.Sprintf(
			"           0x%s: 0x%s (base58: %s) ('%s': '%s')\n",
			slice(hex.EncodeToString([]byte(key))),
			slice(hex.EncodeToString(m[key])),
			slice(base58.Encode(m[key])),
			string(key),
			string(m[key]),
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
func (m kvmap) ForEach(fun func(key kv.Key, value []byte) bool) {
	for k, v := range m {
		if !fun(k, v) {
			return // abort when callback returns false
		}
	}
}

func (m kvmap) ForEachDeterministic(fun func(key kv.Key, value []byte) bool) {
	if m == nil {
		return
	}
	for _, k := range m.sortedKeys() {
		if !fun(k, m[k]) {
			return // abort when callback returns false
		}
	}
}

func (m kvmap) IsEmpty() bool {
	return len(m) == 0
}

func (m kvmap) Set(key kv.Key, value []byte) {
	if value == nil {
		panic("cannot Set(key, nil), use Del() to remove a key/value")
	}
	m[key] = value
}

func (m kvmap) Del(key kv.Key) {
	delete(m, key)
}

func (m kvmap) Has(key kv.Key) (bool, error) {
	_, ok := m[key]
	return ok, nil
}

func (m kvmap) Iterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) error {
	for k, v := range m {
		if !k.HasPrefix(prefix) {
			continue
		}
		if !f(k, v) {
			break
		}
	}
	return nil
}

func (m kvmap) IterateKeys(prefix kv.Key, f func(key kv.Key) bool) error {
	for k, _ := range m {
		if !k.HasPrefix(prefix) {
			continue
		}
		if !f(k) {
			break
		}
	}
	return nil
}

func (m kvmap) Get(key kv.Key) ([]byte, error) {
	v, _ := m[key]
	return v, nil
}

func (m kvmap) Write(w io.Writer) error {
	keys := m.sortedKeys()
	if err := util.WriteUint64(w, uint64(len(keys))); err != nil {
		return err
	}
	for _, k := range keys {
		if err := util.WriteBytes16(w, []byte(k)); err != nil {
			return err
		}
		if err := util.WriteBytes32(w, m[k]); err != nil {
			return err
		}
	}
	return nil
}

func (m kvmap) Read(r io.Reader) error {
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
		m.Set(kv.Key(k), v)
	}
	return nil
}
