package sandbox

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
)

type stateWrapper struct {
	virtualState state.VirtualState
	stateUpdate  state.StateUpdate
}

func (s *stateWrapper) MustCodec() kv.MustCodec {
	return kv.NewMustCodec(s)
}

func (s *stateWrapper) Has(name kv.Key) (bool, error) {
	mut := s.stateUpdate.Mutations().Latest(name)
	if mut != nil {
		return mut.Value() != nil, nil
	}
	return s.virtualState.Variables().Has(name)
}

func (s *stateWrapper) Get(name kv.Key) ([]byte, error) {
	mut := s.stateUpdate.Mutations().Latest(name)
	if mut != nil {
		return mut.Value(), nil
	}
	return s.virtualState.Variables().Get(name)
}

func (s *stateWrapper) Del(name kv.Key) {
	s.stateUpdate.Mutations().Add(kv.NewMutationDel(name))
}

func (s *stateWrapper) Set(name kv.Key, value []byte) {
	s.stateUpdate.Mutations().Add(kv.NewMutationSet(name, value))
}
