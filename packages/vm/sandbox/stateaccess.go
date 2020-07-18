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

func (s *stateWrapper) findLatestMutation(name kv.Key) *kv.Mutation {
	// FIXME: this is O(N) with N = amount of accumulated mutations
	// it could be improved by caching the latest mutation for evey key
	muts := s.stateUpdate.Mutations()
	for i := muts.Len() - 1; i >= 0; i-- {
		m := muts.At(i)
		if (*m).Key() == name {
			// The key-value pair has been modified during the current request
			// return the latest assigned value
			return m
		}
	}
	// The key-value pair has not been modified
	// Fetch its value from the virtual state
	return nil
}

func (s *stateWrapper) Has(name kv.Key) (bool, error) {
	mut := s.findLatestMutation(name)
	if mut != nil {
		return (*mut).Value() != nil, nil
	}
	return s.virtualState.Variables().Has(name)
}

func (s *stateWrapper) Get(name kv.Key) ([]byte, error) {
	mut := s.findLatestMutation(name)
	if mut != nil {
		return (*mut).Value(), nil
	}
	return s.virtualState.Variables().Get(name)
}

func (s *stateWrapper) Del(name kv.Key) {
	s.stateUpdate.Mutations().Add(kv.NewMutationDel(name))
}

func (s *stateWrapper) Set(name kv.Key, value []byte) {
	s.stateUpdate.Mutations().Add(kv.NewMutationSet(name, value))
}
