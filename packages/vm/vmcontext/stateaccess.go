package vmcontext

import (
	"sort"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
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
		vmctx.currentStateUpdate,
	)
}

func (s stateWrapper) Has(name kv.Key) (bool, error) {
	name = s.addContractSubPartition(name)
	_, ok := s.stateUpdate.Mutations().Sets[name]
	if ok {
		return true, nil
	}
	return s.virtualState.KVStore().Has(name)
}

func (s stateWrapper) Iterate(prefix kv.Key, f func(kv.Key, []byte) bool) error {
	var err error
	err2 := s.IterateKeys(prefix, func(k kv.Key) bool {
		var v []byte
		v, err = s.Get(k)
		if err != nil {
			return false
		}
		return f(k, v)
	})
	if err2 != nil {
		return err2
	}
	return err
}

func (s stateWrapper) IterateKeys(prefix kv.Key, f func(key kv.Key) bool) error {
	prefix = s.addContractSubPartition(prefix)
	for k := range s.stateUpdate.Mutations().Sets {
		if k.HasPrefix(prefix) {
			if !f(k[len(s.contractSubPartitionPrefix):]) {
				return nil
			}
		}
	}
	return s.virtualState.KVStore().IterateKeys(prefix, func(k kv.Key) bool {
		if !s.stateUpdate.Mutations().Contains(k) {
			return f(k[len(s.contractSubPartitionPrefix):])
		}
		return true
	})
}

func (s stateWrapper) IterateSorted(prefix kv.Key, f func(kv.Key, []byte) bool) error {
	var err error
	err2 := s.IterateKeysSorted(prefix, func(k kv.Key) bool {
		var v []byte
		v, err = s.Get(k)
		if err != nil {
			return false
		}
		return f(k, v)
	})
	if err2 != nil {
		return err2
	}
	return err
}

func (s stateWrapper) IterateKeysSorted(prefix kv.Key, f func(key kv.Key) bool) error {
	prefix = s.addContractSubPartition(prefix)
	var keys []kv.Key
	for k := range s.stateUpdate.Mutations().Sets {
		if k.HasPrefix(prefix) {
			keys = append(keys, k)
		}
	}
	err := s.virtualState.KVStore().IterateKeysSorted(prefix, func(k kv.Key) bool {
		if !s.stateUpdate.Mutations().Contains(k) {
			keys = append(keys, k)
		}
		return true
	})
	if err != nil {
		return err
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	for _, k := range keys {
		if !f(k[len(s.contractSubPartitionPrefix):]) {
			break
		}
	}
	return nil
}

func (s stateWrapper) Get(name kv.Key) ([]byte, error) {
	name = s.addContractSubPartition(name)
	v, ok := s.stateUpdate.Mutations().Sets[name]
	if ok {
		return v, nil
	}
	return s.virtualState.KVStore().Get(name)
}

func (s stateWrapper) Del(name kv.Key) {
	name = s.addContractSubPartition(name)
	s.stateUpdate.Mutations().Del(name)
}

func (s stateWrapper) Set(name kv.Key, value []byte) {
	name = s.addContractSubPartition(name)
	s.stateUpdate.Mutations().Set(name, value)
}

func (vmctx *VMContext) State() kv.KVStore {
	w := vmctx.stateWrapper()
	// vmctx.log.Debugf("state wrapper: %s", w.contractHname.String())
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

func (s stateWrapper) MustIterateSorted(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	kv.MustIterateSorted(s, prefix, f)
}

func (s stateWrapper) MustIterateKeysSorted(prefix kv.Key, f func(key kv.Key) bool) {
	kv.MustIterateKeysSorted(s, prefix, f)
}
