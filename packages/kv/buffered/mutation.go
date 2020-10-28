package buffered

import (
	"fmt"
	"io"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
)

// Mutation represents a single "set" or "del" operation over a KVStore
type Mutation interface {
	Read(io.Reader) error
	Write(io.Writer) error

	String() string

	ApplyTo(kv kv.KVStore)

	// Key returns the key that is mutated
	Key() kv.Key
	// Value returns the value after the mutation (nil if deleted)
	Value() []byte

	getMagic() int
}

type MutationSequence interface {
	Read(io.Reader) error
	Write(io.Writer) error

	String() string

	Clone() MutationSequence

	Len() int

	// Iterate over all mutations in order, even ones affecting the same key repeatedly
	Iterate(func(mut Mutation) bool)
	// Iterate over the latest mutation recorded for each key
	IterateLatest(func(key kv.Key, mut Mutation) bool)
	// Iterate over the latest value recorded for each non-deleted key
	IterateValues(prefix kv.Key, f func(key kv.Key, value []byte) bool) (map[kv.Key]bool, bool)

	Latest(key kv.Key) Mutation

	Add(mut Mutation)

	ApplyTo(kv kv.KVStore)
}

const (
	mutationMagicSet = iota
	mutationMagicDel
)

type mutationSequence struct {
	muts        []Mutation
	latestByKey map[kv.Key]*Mutation
}

func NewMutationSequence() MutationSequence {
	return &mutationSequence{
		muts:        make([]Mutation, 0),
		latestByKey: make(map[kv.Key]*Mutation),
	}
}

func (ms *mutationSequence) String() string {
	ret := ""
	for _, mut := range ms.muts {
		ret += fmt.Sprintf("[%s] ", mut.String())
	}
	return ret
}

func (ms *mutationSequence) Write(w io.Writer) error {
	n := len(ms.muts)
	if n > util.MaxUint16 {
		return fmt.Errorf("Too many mutations")
	}
	if err := util.WriteUint16(w, uint16(n)); err != nil {
		return err
	}
	for _, mut := range ms.muts {
		if err := util.WriteUint16(w, uint16(mut.getMagic())); err != nil {
			return err
		}
		if err := mut.Write(w); err != nil {
			return err
		}
	}
	return nil
}

func (ms *mutationSequence) Read(r io.Reader) error {
	var n uint16
	if err := util.ReadUint16(r, &n); err != nil {
		return err
	}
	for i := uint16(0); i < n; i++ {
		var magic uint16
		if err := util.ReadUint16(r, &magic); err != nil {
			return err
		}
		mut, err := newFromMagic(int(magic))
		if err != nil {
			return err
		}
		if err = mut.Read(r); err != nil {
			return err
		}
		ms.Add(mut)
	}
	return nil
}

func (ms *mutationSequence) Iterate(f func(mut Mutation) bool) {
	for _, mut := range ms.muts {
		if !f(mut) {
			break
		}
	}
}

func (ms *mutationSequence) IterateLatest(f func(kv.Key, Mutation) bool) {
	for key, mut := range ms.latestByKey {
		if !f(key, *mut) {
			break
		}
	}
}

func (ms *mutationSequence) IterateValues(prefix kv.Key, f func(key kv.Key, value []byte) bool) (map[kv.Key]bool, bool) {
	seen := make(map[kv.Key]bool)
	for key, mut := range ms.latestByKey {
		if !key.HasPrefix(prefix) {
			continue
		}
		seen[key] = true
		v := (*mut).Value()
		if v != nil && !f(key, v) {
			return seen, true
		}
	}
	return seen, false
}

func (ms *mutationSequence) Len() int {
	return len(ms.muts)
}

func (ms *mutationSequence) Add(mut Mutation) {
	ms.muts = append(ms.muts, mut)
	ms.latestByKey[mut.Key()] = &mut
}

func (ms *mutationSequence) ApplyTo(kv kv.KVStore) {
	for _, mut := range ms.muts {
		mut.ApplyTo(kv)
	}
}

func (ms *mutationSequence) Latest(key kv.Key) Mutation {
	mut, ok := ms.latestByKey[key]
	if !ok {
		return nil
	}
	return *mut
}

func (ms *mutationSequence) Clone() MutationSequence {
	mapClone := make(map[kv.Key]*Mutation)
	for k, v := range ms.latestByKey {
		mapClone[k] = v
	}
	return &mutationSequence{muts: ms.muts[:], latestByKey: mapClone}
}

type mutationSet struct {
	k kv.Key
	v []byte
}

type mutationDel struct {
	k kv.Key
}

func newFromMagic(magic int) (Mutation, error) {
	switch magic {
	case mutationMagicSet:
		return &mutationSet{}, nil
	case mutationMagicDel:
		return &mutationDel{}, nil
	}
	return nil, fmt.Errorf("Unknown mutation magic %d", magic)
}

func NewMutationSet(k kv.Key, v []byte) *mutationSet {
	return &mutationSet{k: k, v: v}
}

func (m *mutationSet) getMagic() int {
	return mutationMagicSet
}

func (m *mutationSet) Write(w io.Writer) error {
	if err := util.WriteBytes16(w, []byte(m.k)); err != nil {
		return err
	}
	if err := util.WriteBytes32(w, m.v); err != nil {
		return err
	}
	return nil
}

func (m *mutationSet) Read(r io.Reader) error {
	k, err := util.ReadBytes16(r)
	if err != nil {
		return err
	}
	v, err := util.ReadBytes32(r)
	if err != nil {
		return err
	}
	m.k = kv.Key(k)
	m.v = v
	return nil
}

func (m *mutationSet) String() string {
	return fmt.Sprintf("SET \"%s\"={%x}", m.k, m.v)
}

func (m *mutationSet) Key() kv.Key {
	return m.k
}

func (m *mutationSet) Value() []byte {
	return m.v
}

func (m *mutationSet) ApplyTo(kv kv.KVStore) {
	kv.Set(m.k, m.v)
}

func (m *mutationDel) getMagic() int {
	return mutationMagicDel
}

func NewMutationDel(k kv.Key) *mutationDel {
	return &mutationDel{k: k}
}

func (m *mutationDel) Write(w io.Writer) error {
	return util.WriteBytes16(w, []byte(m.k))
}

func (m *mutationDel) Read(r io.Reader) error {
	k, err := util.ReadBytes16(r)
	if err != nil {
		return err
	}
	m.k = kv.Key(k)
	return nil
}

func (m *mutationDel) String() string {
	return fmt.Sprintf("DEL %s", m.k)
}

func (m *mutationDel) Key() kv.Key {
	return m.k
}

func (m *mutationDel) Value() []byte {
	return nil
}

func (m *mutationDel) ApplyTo(kv kv.KVStore) {
	kv.Del(m.k)
}
