package vmcontext

import (
	"bytes"
	"io"

	"github.com/iotaledger/wasp/packages/kv/buffered"
)

type StateUpdate struct {
	Mutations *buffered.Mutations
}

// NewStateUpdate creates a state update with timestamp mutation, if provided
func NewStateUpdate() *StateUpdate {
	return &StateUpdate{
		Mutations: buffered.NewMutations(),
	}
}

func newStateUpdateFromReader(r io.Reader) (*StateUpdate, error) {
	ret := &StateUpdate{
		Mutations: buffered.NewMutations(),
	}
	err := ret.Read(r)
	return ret, err
}

func (su *StateUpdate) Clone() *StateUpdate {
	return &StateUpdate{Mutations: su.Mutations.Clone()}
}

func (su *StateUpdate) Bytes() []byte {
	var buf bytes.Buffer
	_ = su.Write(&buf)
	return buf.Bytes()
}

func (su *StateUpdate) Write(w io.Writer) error {
	return su.Mutations.Write(w)
}

func (su *StateUpdate) Read(r io.Reader) error {
	return su.Mutations.Read(r)
}
