package vmcontext

import (
	"github.com/iotaledger/wasp/packages/coret"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/state"
)

type stateWrapper struct {
	contractHname              coret.Hname
	contractSubPartitionPrefix kv.Key
	virtualState               state.VirtualState
	stateUpdate                state.StateUpdate
}

func newStateWrapper(contractHname coret.Hname, virtualState state.VirtualState, stateUpdate state.StateUpdate) stateWrapper {
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

// ------------------------ VMContext kvstore implementation

func (vmctx *VMContext) Set(name kv.Key, value []byte) {
	vmctx.stateWrapper().Set(name, value)
}

func (vmctx *VMContext) Has(name kv.Key) (bool, error) {
	return vmctx.stateWrapper().Has(name)
}

func (vmctx *VMContext) Iterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) error {
	return vmctx.stateWrapper().Iterate(prefix, f)
}

func (vmctx *VMContext) IterateKeys(prefix kv.Key, f func(key kv.Key) bool) error {
	return vmctx.stateWrapper().IterateKeys(prefix, f)
}

func (vmctx *VMContext) Get(name kv.Key) ([]byte, error) {
	return vmctx.stateWrapper().Get(name)
}

func (vmctx *VMContext) Del(name kv.Key) {
	vmctx.stateWrapper().Del(name)
}

func (vmctx *VMContext) State() codec.MutableMustCodec {
	w := vmctx.stateWrapper()
	vmctx.log.Debugf("state wrapper: %s", w.contractHname.String())
	return codec.NewMustCodec(w)
}
