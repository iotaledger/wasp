package kv

import (
	"fmt"
	"io"

	"github.com/iotaledger/wasp/packages/util"
)

// Mutation represents a single "set" or "del" operation over a KVStore
type Mutation interface {
	Read(io.Reader) error
	Write(io.Writer) error

	String() string

	ApplyTo(kv KVStore)

	// Key returns the key that is mutated
	Key() Key
	// Value returns the value after the mutation (nil if deleted)
	Value() []byte

	getMagic() int
}

type MutationSequence interface {
	Read(io.Reader) error
	Write(io.Writer) error

	String() string

	ForEach(func(mut Mutation))
	At(i int) *Mutation
	Len() int

	Add(mut Mutation)
	AddAll(ms MutationSequence)

	ApplyTo(kv KVStore)
}

// MutationMap stores the latest mutation applied to each key
type MutationMap interface {
	Get(key Key) Mutation
	Add(mut Mutation)
	Clone() MutationMap
	Iterate(func(Key, Mutation) bool)
}

const (
	mutationMagicSet = iota
	mutationMagicDel
)

type mutationSequence struct {
	muts []Mutation
}

func NewMutationSequence() MutationSequence {
	return &mutationSequence{muts: make([]Mutation, 0)}
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

func (ms *mutationSequence) ForEach(f func(mut Mutation)) {
	for _, mut := range ms.muts {
		f(mut)
	}
}

func (ms *mutationSequence) Len() int {
	return len(ms.muts)
}

func (ms *mutationSequence) At(i int) *Mutation {
	return &ms.muts[i]
}

func (ms *mutationSequence) Add(mut Mutation) {
	ms.muts = append(ms.muts, mut)
}

func (ms *mutationSequence) AddAll(other MutationSequence) {
	other.ForEach(func(mut Mutation) {
		ms.Add(mut)
	})
}

func (ms *mutationSequence) ApplyTo(kv KVStore) {
	for _, mut := range ms.muts {
		mut.ApplyTo(kv)
	}
}

type mutationMap map[Key]Mutation

func NewMutationMap() MutationMap {
	return make(mutationMap)
}

func (m mutationMap) Get(key Key) Mutation {
	v, _ := m[key]
	return v
}

func (m mutationMap) Add(mut Mutation) {
	m[mut.Key()] = mut
}

func (m mutationMap) Iterate(f func(Key, Mutation) bool) {
	for k, v := range m {
		if !f(k, v) {
			break
		}
	}
}

func (m mutationMap) Clone() MutationMap {
	clone := make(mutationMap)
	for k, v := range m {
		clone[k] = v
	}
	return clone
}

type mutationSet struct {
	k Key
	v []byte
}

type mutationDel struct {
	k Key
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

func NewMutationSet(k Key, v []byte) *mutationSet {
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
	m.k = Key(k)
	m.v = v
	return nil
}

func (m *mutationSet) String() string {
	return fmt.Sprintf("SET \"%s\"={%x}", m.k, m.v)
}

func (m *mutationSet) Key() Key {
	return m.k
}

func (m *mutationSet) Value() []byte {
	return m.v
}

func (m *mutationSet) ApplyTo(kv KVStore) {
	kv.Set(m.k, m.v)
}

func (m *mutationDel) getMagic() int {
	return mutationMagicDel
}

func NewMutationDel(k Key) *mutationDel {
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
	m.k = Key(k)
	return nil
}

func (m *mutationDel) String() string {
	return fmt.Sprintf("DEL %s", m.k)
}

func (m *mutationDel) Key() Key {
	return m.k
}

func (m *mutationDel) Value() []byte {
	return nil
}

func (m *mutationDel) ApplyTo(kv KVStore) {
	kv.Del(m.k)
}
