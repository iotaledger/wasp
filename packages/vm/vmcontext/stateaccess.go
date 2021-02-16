package vmcontext

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/state"
)

type stateWrapper struct {
	contractHname              coretypes.Hname
	contractSubPartitionPrefix kv.Key
	virtualState               state.VirtualState
	stateUpdate                state.StateUpdate
}

func newStateWrapper(contractHname coretypes.Hname, virtualState state.VirtualState, stateUpdate state.StateUpdate) stateWrapper {
	return stateWrapper{
		contractHname:              contractHname,
		contractSubPartitionPrefix: kv.Key(contractHname.Bytes()),
		virtualState:               virtualState,
		stateUpdate:                stateUpdate,
	}
}

func (s *stateWrapper) addContractSubPartition(key kv.Key) kv.Key {
	return s.contractSubPartitionPrefix + key
}

func (vmctx *VMContext) stateWrapper() stateWrapper {
	return newStateWrapper(
		vmctx.CurrentContractHname(),
		vmctx.virtualState,
		vmctx.stateUpdate,
	)
}

func (s stateWrapper) Has(name kv.Key) (bool, error) {
	name = s.addContractSubPartition(name)
	mut := s.stateUpdate.Mutations().Latest(name)
	if mut != nil {
		return mut.Value() != nil, nil
	}
	return s.virtualState.Variables().Has(name)
}

func (s stateWrapper) Iterate(prefix kv.Key, f func(kv.Key, []byte) bool) error {
	prefix = s.addContractSubPartition(prefix)
	seen, done := s.stateUpdate.Mutations().IterateValues(prefix, func(key kv.Key, value []byte) bool {
		return f(key[len(s.contractSubPartitionPrefix):], value)
	})
	if done {
		return nil
	}
	return s.virtualState.Variables().Iterate(prefix, func(key kv.Key, value []byte) bool {
		_, ok := seen[key]
		if ok {
			return true
		}
		return f(key[len(s.contractSubPartitionPrefix):], value)
	})
}

func (s stateWrapper) IterateKeys(prefix kv.Key, f func(key kv.Key) bool) error {
	prefix = s.addContractSubPartition(prefix)
	seen, done := s.stateUpdate.Mutations().IterateValues(prefix, func(key kv.Key, value []byte) bool {
		return f(key[len(s.contractSubPartitionPrefix):])
	})
	if done {
		return nil
	}
	return s.virtualState.Variables().IterateKeys(prefix, func(key kv.Key) bool {
		_, ok := seen[key]
		if ok {
			return true
		}
		return f(key[len(s.contractSubPartitionPrefix):])
	})
}

func (s stateWrapper) Get(name kv.Key) ([]byte, error) {
	name = s.addContractSubPartition(name)
	mut := s.stateUpdate.Mutations().Latest(name)
	if mut != nil {
		return mut.Value(), nil
	}
	return s.virtualState.Variables().Get(name)
}

func (s stateWrapper) Del(name kv.Key) {
	name = s.addContractSubPartition(name)
	s.stateUpdate.Mutations().Add(buffered.NewMutationDel(name))
}

func (s stateWrapper) Set(name kv.Key, value []byte) {
	name = s.addContractSubPartition(name)
	s.stateUpdate.Mutations().Add(buffered.NewMutationSet(name, value))
}

func (vmctx *VMContext) State() kv.KVStore {
	w := vmctx.stateWrapper()
	//vmctx.log.Debugf("state wrapper: %s", w.contractHname.String())
	return w
}

func (s stateWrapper) MustGet(key kv.Key) []byte {
	return kv.MustGet(s, key)
}

func (s stateWrapper) MustHas(key kv.Key) bool {
	return kv.MustHas(s, key)
}

func (s stateWrapper) MustIterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	kv.MustIterate(s, prefix, f)
}

func (s stateWrapper) MustIterateKeys(prefix kv.Key, f func(key kv.Key) bool) {
	kv.MustIterateKeys(s, prefix, f)
}
