package table

import (
	"encoding/hex"
	"fmt"
	"github.com/mr-tron/base58"
	"io"
	"sort"

	"github.com/iotaledger/wasp/packages/util"
)

// MemTable is a Table backed by an in-memory map
type MemTable interface {
	Table

	IsEmpty() bool

	ToMap() map[Key][]byte

	ForEach(func(key Key, value []byte) bool)
	ForEachDeterministic(func(key Key, value []byte) bool)
	Clone() MemTable

	Read(io.Reader) error
	Write(io.Writer) error

	String() string

	Mutations() MutationSequence

	Codec() Codec
}

type memtable map[Key][]byte

// create/clone
func NewMemTable() MemTable {
	return make(memtable)
}

func (vr memtable) Clone() MemTable {
	clone := make(memtable)
	vr.ForEach(func(key Key, value []byte) bool {
		clone.Set(key, value)
		return true
	})
	return clone
}

func FromMap(m map[Key][]byte) MemTable {
	return memtable(m)
}

func (vr memtable) ToMap() map[Key][]byte {
	return memtable(vr)
}

type keySlice []Key

func (a keySlice) Len() int           { return len(a) }
func (a keySlice) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a keySlice) Less(i, j int) bool { return a[i] < a[j] }

func (vr memtable) sortedKeys() []Key {
	keys := make([]Key, 0)
	for k := range vr {
		keys = append(keys, k)
	}
	sort.Sort(keySlice(keys))
	return keys
}

func (vr memtable) String() string {
	ret := "         MemTable:\n"
	for _, key := range vr.sortedKeys() {
		ret += fmt.Sprintf(
			"           0x%s: 0x%s (base58: %s)\n",
			slice(hex.EncodeToString([]byte(key))),
			slice(hex.EncodeToString(vr[key])),
			slice(base58.Encode(vr[key])),
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

func (vr memtable) Codec() Codec {
	return NewCodec(vr)
}

func (vr memtable) Mutations() MutationSequence {
	ms := NewMutationSequence()
	vr.ForEachDeterministic(func(key Key, value []byte) bool {
		ms.Add(NewMutationSet(key, value))
		return true
	})
	return ms
}

// NON DETERMINISTIC!
func (vr memtable) ForEach(fun func(key Key, value []byte) bool) {
	for k, v := range vr {
		if !fun(k, v) {
			return // abort when callback returns false
		}
	}
}

func (vr memtable) ForEachDeterministic(fun func(key Key, value []byte) bool) {
	if vr == nil {
		return
	}
	for _, k := range vr.sortedKeys() {
		if !fun(k, vr[k]) {
			return // abort when callback returns false
		}
	}
}

func (vr memtable) IsEmpty() bool {
	return len(vr) == 0
}

func (vr memtable) Set(key Key, value []byte) {
	if value == nil {
		panic("cannot Set(key, nil), use Del() to remove a key/value")
	}
	vr[key] = value
}

func (vr memtable) Del(key Key) {
	delete(vr, key)
}

func (vr memtable) Get(key Key) ([]byte, error) {
	v, _ := vr[key]
	return v, nil
}

func (vr memtable) Write(w io.Writer) error {
	keys := vr.sortedKeys()
	if err := util.WriteUint64(w, uint64(len(keys))); err != nil {
		return err
	}
	for _, k := range keys {
		if err := util.WriteBytes16(w, []byte(k)); err != nil {
			return err
		}
		if err := util.WriteBytes32(w, vr[k]); err != nil {
			return err
		}
	}
	return nil
}

func (vr memtable) Read(r io.Reader) error {
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
		vr.Set(Key(k), v)
	}
	return nil
}
