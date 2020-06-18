package variables

import (
	"fmt"
	"io"

	"github.com/iotaledger/wasp/packages/util"
)

type Mutation interface {
	Write(io.Writer) error
	String() string
	Apply(vs Variables)
}

func ReadMutation(r io.Reader) (Mutation, error) {
	var mutationType uint16
	if err := util.ReadUint16(r, &mutationType); err != nil {
		return nil, err
	}
	switch mutationType {
	case mutationTypeSet:
		return readMutationSet(r)
	case mutationTypeDel:
		return readMutationDel(r)
	}
	return nil, fmt.Errorf("Unknown mutation type %d", mutationType)
}

type MutationSet struct {
	k string
	v []byte
}

type MutationDel struct {
	k string
}

const (
	mutationTypeSet = iota
	mutationTypeDel
)

func NewMutationSet(k string, v []byte) *MutationSet {
	return &MutationSet{k: k, v: v}
}

func (m *MutationSet) Write(w io.Writer) error {
	if err := util.WriteUint16(w, mutationTypeSet); err != nil {
		return err
	}
	if err := util.WriteString16(w, m.k); err != nil {
		return err
	}
	if err := util.WriteBytes16(w, m.v); err != nil {
		return err
	}
	return nil
}

func readMutationSet(r io.Reader) (*MutationSet, error) {
	k, err := util.ReadString16(r)
	if err != nil {
		return nil, err
	}
	v, err := util.ReadBytes16(r)
	if err != nil {
		return nil, err
	}
	return &MutationSet{k: k, v: v}, nil
}

func (m *MutationSet) String() string {
	return fmt.Sprintf("SET %s=%v", m.k, m.v)
}

func (m *MutationSet) Apply(vs Variables) {
	vs.Set(m.k, m.v)
}

func NewMutationDel(k string) *MutationDel {
	return &MutationDel{k: k}
}

func (m *MutationDel) Write(w io.Writer) error {
	if err := util.WriteUint16(w, mutationTypeDel); err != nil {
		return err
	}
	if err := util.WriteString16(w, m.k); err != nil {
		return err
	}
	return nil
}

func readMutationDel(r io.Reader) (*MutationDel, error) {
	k, err := util.ReadString16(r)
	if err != nil {
		return nil, err
	}
	return &MutationDel{k: k}, nil
}

func (m *MutationDel) String() string {
	return fmt.Sprintf("DEL %s", m.k)
}

func (m *MutationDel) Apply(vs Variables) {
	vs.Del(m.k)
}

