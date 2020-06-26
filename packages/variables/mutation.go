package variables

import (
	"fmt"
	"io"

	"github.com/iotaledger/wasp/packages/util"
)

// Mutation represents a single operation over a Variables key-value
type Mutation interface {
	Read(io.Reader) error
	Write(io.Writer) error

	String() string

	ApplyTo(vs Variables)

	Key() string
	Value() (v []byte, ok bool)

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

	ApplyTo(vs Variables)
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

func (ms *mutationSequence) ApplyTo(v Variables) {
	for _, mut := range ms.muts {
		mut.ApplyTo(v)
	}
}

type mutationSet struct {
	k string
	v []byte
}

type mutationDel struct {
	k string
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

func NewMutationSet(k string, v []byte) *mutationSet {
	return &mutationSet{k: k, v: v}
}

func (m *mutationSet) getMagic() int {
	return mutationMagicSet
}

func (m *mutationSet) Write(w io.Writer) error {
	if err := util.WriteString16(w, m.k); err != nil {
		return err
	}
	if err := util.WriteBytes32(w, m.v); err != nil {
		return err
	}
	return nil
}

func (m *mutationSet) Read(r io.Reader) error {
	k, err := util.ReadString16(r)
	if err != nil {
		return err
	}
	v, err := util.ReadBytes32(r)
	if err != nil {
		return err
	}
	m.k = k
	m.v = v
	return nil
}

func (m *mutationSet) String() string {
	return fmt.Sprintf("SET \"%s\"={%x}", m.k, m.v)
}

func (m *mutationSet) Key() string {
	return m.k
}

func (m *mutationSet) Value() (v []byte, ok bool) {
	return m.v, true
}

func (m *mutationSet) ApplyTo(vs Variables) {
	vs.Set(m.k, m.v)
}

func (m *mutationDel) getMagic() int {
	return mutationMagicDel
}

func NewMutationDel(k string) *mutationDel {
	return &mutationDel{k: k}
}

func (m *mutationDel) Write(w io.Writer) error {
	return util.WriteString16(w, m.k)
}

func (m *mutationDel) Read(r io.Reader) error {
	k, err := util.ReadString16(r)
	if err != nil {
		return err
	}
	m.k = k
	return nil
}

func (m *mutationDel) String() string {
	return fmt.Sprintf("DEL %s", m.k)
}

func (m *mutationDel) Key() string {
	return m.k
}

func (m *mutationDel) Value() (v []byte, ok bool) {
	return nil, false
}

func (m *mutationDel) ApplyTo(vs Variables) {
	vs.Del(m.k)
}
