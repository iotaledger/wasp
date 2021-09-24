package buffered

import (
	"bytes"
	"io"
	"sort"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
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

func (ms *Mutations) Bytes() []byte {
	var buf bytes.Buffer
	_ = ms.Write(&buf)
	return buf.Bytes()
}

func (ms *Mutations) Write(w io.Writer) error {
	if err := util.WriteUint32(w, uint32(len(ms.Sets))); err != nil {
		return err
	}
	for _, item := range ms.SetsSorted() {
		if err := util.WriteString16(w, string(item.Key)); err != nil {
			return err
		}
		if err := util.WriteBytes32(w, item.Value); err != nil {
			return err
		}
	}
	if err := util.WriteUint32(w, uint32(len(ms.Dels))); err != nil {
		return err
	}
	for _, k := range ms.DelsSorted() {
		if err := util.WriteString16(w, string(k)); err != nil {
			return err
		}
	}
	return nil
}

//nolint:gocritic
func (ms *Mutations) Read(r io.Reader) error {
	var err error
	var n uint32
	if err = util.ReadUint32(r, &n); err != nil {
		return err
	}
	for i := uint32(0); i < n; i++ {
		var k string
		var v []byte
		if k, err = util.ReadString16(r); err != nil {
			return err
		}
		if v, err = util.ReadBytes32(r); err != nil {
			return err
		}
		ms.Set(kv.Key(k), v)
	}
	if err = util.ReadUint32(r, &n); err != nil {
		return err
	}
	for i := uint32(0); i < n; i++ {
		var k string
		if k, err = util.ReadString16(r); err != nil {
			return err
		}
		ms.Del(kv.Key(k))
	}
	return nil
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
		clone.Set(k, v)
	}
	for k := range ms.Dels {
		clone.Del(k)
	}
	// left unlocked intentionally
	return clone
}

func (ms *Mutations) IsEmpty() bool {
	return len(ms.Sets) == 0 && len(ms.Dels) == 0
}
