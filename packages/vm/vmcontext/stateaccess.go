package vmcontext

import (
	"sort"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
)

type chainStateWrapper struct {
	vmctx *VMContext
}

// chainState is a KVStore to access the whole state of the chain. To access the partition of the contract
// use subrealm or methid (vmctx *VMContext) State() kv.KVStore
// Each access to the state checks validity of the state. It panics optimistic reader error
// if state has been invalidated (new state was committed)
func (vmctx *VMContext) chainState() chainStateWrapper {
	return chainStateWrapper{vmctx}
}

func (s chainStateWrapper) Has(name kv.Key) (bool, error) {
	s.vmctx.task.SolidStateBaseline.MustValidate()

	if _, ok := s.vmctx.currentStateUpdate.Mutations().Sets[name]; ok {
		return true, nil
	}
	return s.vmctx.virtualState.KVStore().Has(name)
}

func (s chainStateWrapper) Iterate(prefix kv.Key, f func(kv.Key, []byte) bool) error {
	s.vmctx.task.SolidStateBaseline.MustValidate()

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

func (s chainStateWrapper) IterateKeys(prefix kv.Key, f func(key kv.Key) bool) error {
	s.vmctx.task.SolidStateBaseline.MustValidate()

	for k := range s.vmctx.currentStateUpdate.Mutations().Sets {
		if k.HasPrefix(prefix) {
			if !f(k) {
				return nil
			}
		}
	}
	return s.vmctx.virtualState.KVStore().IterateKeys(prefix, func(k kv.Key) bool {
		if !s.vmctx.currentStateUpdate.Mutations().Contains(k) {
			return f(k)
		}
		return true
	})
}

func (s chainStateWrapper) IterateSorted(prefix kv.Key, f func(kv.Key, []byte) bool) error {
	s.vmctx.task.SolidStateBaseline.MustValidate()

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

func (s chainStateWrapper) IterateKeysSorted(prefix kv.Key, f func(key kv.Key) bool) error {
	s.vmctx.task.SolidStateBaseline.MustValidate()

	var keys []kv.Key
	for k := range s.vmctx.currentStateUpdate.Mutations().Sets {
		if k.HasPrefix(prefix) {
			keys = append(keys, k)
		}
	}
	err := s.vmctx.virtualState.KVStore().IterateKeysSorted(prefix, func(k kv.Key) bool {
		if !s.vmctx.currentStateUpdate.Mutations().Contains(k) {
			keys = append(keys, k)
		}
		return true
	})
	if err != nil {
		return err
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	for _, k := range keys {
		if !f(k) {
			break
		}
	}
	return nil
}

func (s chainStateWrapper) Get(name kv.Key) ([]byte, error) {
	s.vmctx.task.SolidStateBaseline.MustValidate()

	v, ok := s.vmctx.currentStateUpdate.Mutations().Sets[name]
	if ok {
		return v, nil
	}
	return s.vmctx.virtualState.KVStore().Get(name)
}

func (s chainStateWrapper) Del(name kv.Key) {
	s.vmctx.task.SolidStateBaseline.MustValidate()

	s.vmctx.currentStateUpdate.Mutations().Del(name)
}

func (s chainStateWrapper) Set(name kv.Key, value []byte) {
	s.vmctx.task.SolidStateBaseline.MustValidate()

	s.vmctx.currentStateUpdate.Mutations().Set(name, value)
}

func (vmctx *VMContext) State() kv.KVStore {
	vmctx.task.SolidStateBaseline.MustValidate()
	return subrealm.New(vmctx.chainState(), kv.Key(vmctx.CurrentContractHname().Bytes()))
}

// Disabled because of recursive calls
//func (vmctx *VMContext) State(burnGas ...kv.BurnGasFn) kv.KVStore {
//	vmctx.task.SolidStateBaseline.MustValidate()
//	store := subrealm.New(vmctx.chainState(), kv.Key(vmctx.CurrentContractHname().Bytes()))
//	if len(burnGas) > 0 {
//		return kv.WithGas(store, burnGas[0])
//	}
//	return store
//}

func (s chainStateWrapper) MustGet(key kv.Key) []byte {
	return kv.MustGet(s, key)
}

func (s chainStateWrapper) MustHas(key kv.Key) bool {
	return kv.MustHas(s, key)
}

func (s chainStateWrapper) MustIterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	kv.MustIterate(s, prefix, f)
}

func (s chainStateWrapper) MustIterateKeys(prefix kv.Key, f func(key kv.Key) bool) {
	kv.MustIterateKeys(s, prefix, f)
}

func (s chainStateWrapper) MustIterateSorted(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	kv.MustIterateSorted(s, prefix, f)
}

func (s chainStateWrapper) MustIterateKeysSorted(prefix kv.Key, f func(key kv.Key) bool) {
	kv.MustIterateKeysSorted(s, prefix, f)
}
