package buffered

import (
	"fmt"
	"maps"
	"sort"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/kv"
)

// Mutations is a set of mutations: one for each key
// It provides a deterministic serialization
type Mutations struct {
	Sets map[kv.Key][]byte
	Dels map[kv.Key]struct{}
}

func NewMutations() *Mutations {
	return &Mutations{
		Sets: make(map[kv.Key][]byte),
		Dels: make(map[kv.Key]struct{}),
	}
}

func MutationsFromBytes(data []byte) (*Mutations, error) {
	return bcs.Unmarshal[*Mutations](data)
}

func (ms *Mutations) Bytes() []byte {
	return bcs.MustMarshal(ms)
}

func (ms *Mutations) SetsSorted() kv.Items {
	var ret kv.Items
	for k, v := range ms.Sets {
		ret = append(ret, kv.Item{Key: k, Value: v})
	}
	sort.Sort(ret)
	return ret
}

func (ms *Mutations) DelsSorted() []kv.Key {
	var ret []kv.Key
	for k := range ms.Dels {
		ret = append(ret, k)
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i] < ret[j] })
	return ret
}

func (ms *Mutations) Contains(k kv.Key) bool {
	_, ok := ms.Sets[k]
	if ok {
		return true
	}
	_, ok = ms.Dels[k]
	return ok
}

func (ms *Mutations) Get(k kv.Key) ([]byte, bool) {
	v, ok := ms.Sets[k]
	if ok {
		return v, ok
	}
	_, ok = ms.Dels[k]
	return nil, ok
}

func (ms *Mutations) Set(k kv.Key, v []byte) {
	if v == nil {
		panic("cannot Set(key, nil), use Del() to remove a key/value")
	}
	delete(ms.Dels, k)
	ms.Sets[k] = v
}

func (ms *Mutations) Del(k kv.Key, baseExists bool) {
	delete(ms.Sets, k)
	if baseExists {
		ms.Dels[k] = struct{}{}
	} else {
		delete(ms.Dels, k)
	}
}

func (ms *Mutations) ApplyTo(w kv.KVWriter) {
	for k, v := range ms.Sets {
		w.Set(k, v)
	}
	for k := range ms.Dels {
		w.Del(k)
	}
}

func (ms *Mutations) Clone() *Mutations {
	return &Mutations{
		Sets: maps.Clone(ms.Sets),
		Dels: maps.Clone(ms.Dels),
	}
}

func (ms *Mutations) IsEmpty() bool {
	return len(ms.Sets) == 0 && len(ms.Dels) == 0
}

func (ms *Mutations) Dump() string {
	ret := "\n"
	for _, it := range ms.SetsSorted() {
		ret += it.Format("    SET %-32s : %s\n")
	}
	for _, d := range ms.DelsSorted() {
		ret += fmt.Sprintf("    DEL %-32s\n", d)
	}
	return ret
}
