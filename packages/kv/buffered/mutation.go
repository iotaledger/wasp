package buffered

import (
	"fmt"
	"io"
	"sort"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util/rwutil"
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
	return rwutil.ReadFromBytes(data, NewMutations())
}

func (ms *Mutations) Bytes() []byte {
	return rwutil.WriteToBytes(ms)
}

func (ms *Mutations) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)

	size := rr.ReadSize32()
	for i := 0; i < size; i++ {
		key := rr.ReadString()
		val := rr.ReadBytes()
		if rr.Err != nil {
			return rr.Err
		}
		ms.Set(kv.Key(key), val)
	}

	size = rr.ReadSize32()
	for i := 0; i < size; i++ {
		key := rr.ReadString()
		if rr.Err != nil {
			return rr.Err
		}
		ms.Del(kv.Key(key))
	}
	return rr.Err
}

func (ms *Mutations) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)

	ww.WriteSize32(len(ms.Sets))
	for _, item := range ms.SetsSorted() {
		ww.WriteString(string(item.Key))
		ww.WriteBytes(item.Value)
	}

	ww.WriteSize32(len(ms.Dels))
	for _, k := range ms.DelsSorted() {
		ww.WriteString(string(k))
	}
	return ww.Err
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

func (ms *Mutations) Del(k kv.Key) {
	delete(ms.Sets, k)
	ms.Dels[k] = struct{}{}
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
	clone := NewMutations()

	for k, v := range ms.Sets {
		clone.Set(k, lo.CopySlice(v))
	}

	for k := range ms.Dels {
		clone.Del(k)
	}

	return clone
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
