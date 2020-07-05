package kv

import (
	"encoding/hex"
	"fmt"
	"github.com/mr-tron/base58"
	"io"
	"sort"

	"github.com/iotaledger/wasp/packages/util"
)

// Map is a KVStore backed by an in-memory map
type Map interface {
	KVStore

	IsEmpty() bool

	ToGoMap() map[Key][]byte

	ForEach(func(key Key, value []byte) bool)
	ForEachDeterministic(func(key Key, value []byte) bool)
	Clone() Map

	Read(io.Reader) error
	Write(io.Writer) error

	String() string

	Mutations() MutationSequence

	Codec() Codec
}

type kvmap map[Key][]byte

// create/clone
func NewMap() Map {
	return make(kvmap)
}

func (m kvmap) Clone() Map {
	clone := make(kvmap)
	m.ForEach(func(key Key, value []byte) bool {
		clone.Set(key, value)
		return true
	})
	return clone
}

func FromGoMap(m map[Key][]byte) Map {
	return kvmap(m)
}

func (m kvmap) ToGoMap() map[Key][]byte {
	return m
}

func (m kvmap) sortedKeys() []Key {
	keys := make([]Key, 0)
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return keys
}

func (m kvmap) String() string {
	ret := "         Map:\n"
	for _, key := range m.sortedKeys() {
		ret += fmt.Sprintf(
			"           0x%s: 0x%s (base58: %s)\n",
			slice(hex.EncodeToString([]byte(key))),
			slice(hex.EncodeToString(m[key])),
			slice(base58.Encode(m[key])),
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

func (m kvmap) Codec() Codec {
	return NewCodec(m)
}

func (m kvmap) Mutations() MutationSequence {
	ms := NewMutationSequence()
	m.ForEachDeterministic(func(key Key, value []byte) bool {
		ms.Add(NewMutationSet(key, value))
		return true
	})
	return ms
}

// NON DETERMINISTIC!
func (m kvmap) ForEach(fun func(key Key, value []byte) bool) {
	for k, v := range m {
		if !fun(k, v) {
			return // abort when callback returns false
		}
	}
}

func (m kvmap) ForEachDeterministic(fun func(key Key, value []byte) bool) {
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

func (m kvmap) Set(key Key, value []byte) {
	if value == nil {
		panic("cannot Set(key, nil), use Del() to remove a key/value")
	}
	m[key] = value
}

func (m kvmap) Del(key Key) {
	delete(m, key)
}

func (m kvmap) Get(key Key) ([]byte, error) {
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
		m.Set(Key(k), v)
	}
	return nil
}
