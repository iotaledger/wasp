package vmcontext

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/state"
)

type stateWrapper struct {
	contractHname coretypes.Hname
	virtualState  state.VirtualState
	stateUpdate   state.StateUpdate
}

func (s *stateWrapper) addContractSubPartition(key kv.Key) kv.Key {
	return kv.Key(s.contractHname.Bytes()) + key
}

func (vmctx *VMContext) stateWrapper() stateWrapper {
	return stateWrapper{
		contractHname: vmctx.CurrentContractHname(),
		virtualState:  vmctx.virtualState,
		stateUpdate:   vmctx.stateUpdate,
	}
}

func (s stateWrapper) Has(name kv.Key) (bool, error) {
	name = s.addContractSubPartition(name)
	mut := s.stateUpdate.Mutations().Latest(name)
	if mut != nil {
		return mut.Value() != nil, nil
	}
	return s.virtualState.Variables().Has(name)
}

func (vmctx *VMContext) Has(name kv.Key) (bool, error) {
	return vmctx.stateWrapper().Has(name)
}

func (s stateWrapper) Iterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) error {
	prefix = s.addContractSubPartition(prefix)
	// TODO is it correct?
	seen, done := s.stateUpdate.Mutations().IterateValues(prefix, f)
	if done {
		return nil
	}
	return s.virtualState.Variables().Iterate(prefix, func(key kv.Key, value []byte) bool {
		_, ok := seen[key]
		if ok {
			return true
		}
		return f(key, value)
	})
}

func (vmctx *VMContext) Iterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) error {
	return vmctx.stateWrapper().Iterate(prefix, f)
}

func (s stateWrapper) IterateKeys(prefix kv.Key, f func(key kv.Key) bool) error {
	prefix = s.addContractSubPartition(prefix)
	seen, done := s.stateUpdate.Mutations().IterateValues(prefix, func(key kv.Key, value []byte) bool {
		return f(key)
	})
	if done {
		return nil
	}
	return s.virtualState.Variables().IterateKeys(prefix, func(key kv.Key) bool {
		_, ok := seen[key]
		if ok {
			return true
		}
		return f(key)
	})
}

func (vmctx *VMContext) IterateKeys(prefix kv.Key, f func(key kv.Key) bool) error {
	return vmctx.stateWrapper().IterateKeys(prefix, f)
}

func (s stateWrapper) Get(name kv.Key) ([]byte, error) {
	name = s.addContractSubPartition(name)
	mut := s.stateUpdate.Mutations().Latest(name)
	if mut != nil {
		return mut.Value(), nil
	}
	return s.virtualState.Variables().Get(name)
}

func (vmctx *VMContext) Get(name kv.Key) ([]byte, error) {
	return vmctx.stateWrapper().Get(name)
}

func (s stateWrapper) Del(name kv.Key) {
	name = s.addContractSubPartition(name)
	s.stateUpdate.Mutations().Add(buffered.NewMutationDel(name))
}

func (vmctx *VMContext) Del(name kv.Key) {
	vmctx.stateWrapper().Del(name)
}

func (s stateWrapper) Set(name kv.Key, value []byte) {
	name = s.addContractSubPartition(name)
	s.stateUpdate.Mutations().Add(buffered.NewMutationSet(name, value))
}

func (vmctx *VMContext) Set(name kv.Key, value []byte) {
	vmctx.stateWrapper().Set(name, value)
}
