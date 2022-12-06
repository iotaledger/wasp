package vmcontext

import (
	"sort"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

type chainStateWrapper struct {
	vmctx *VMContext
}

func (vmctx *VMContext) chainState() chainStateWrapper {
	return chainStateWrapper{vmctx}
}

func (s chainStateWrapper) Has(name kv.Key) (bool, error) {
	if _, ok := s.vmctx.currentStateUpdate.Mutations.Sets[name]; ok {
		return true, nil
	}
	if _, wasDeleted := s.vmctx.currentStateUpdate.Mutations.Dels[name]; wasDeleted {
		return false, nil
	}
	return s.vmctx.stateDraft.Has(name)
}

func (s chainStateWrapper) Iterate(prefix kv.Key, f func(kv.Key, []byte) bool) error {
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
	for k := range s.vmctx.currentStateUpdate.Mutations.Sets {
		if k.HasPrefix(prefix) {
			if !f(k) {
				return nil
			}
		}
	}
	return s.vmctx.stateDraft.IterateKeys(prefix, func(k kv.Key) bool {
		if !s.vmctx.currentStateUpdate.Mutations.Contains(k) {
			return f(k)
		}
		return true
	})
}

func (s chainStateWrapper) IterateSorted(prefix kv.Key, f func(kv.Key, []byte) bool) error {
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
	var keys []kv.Key
	for k := range s.vmctx.currentStateUpdate.Mutations.Sets {
		if k.HasPrefix(prefix) {
			keys = append(keys, k)
		}
	}
	err := s.vmctx.stateDraft.IterateKeysSorted(prefix, func(k kv.Key) bool {
		if !s.vmctx.currentStateUpdate.Mutations.Contains(k) {
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
	v, ok := s.vmctx.currentStateUpdate.Mutations.Sets[name]
	if ok {
		return v, nil
	}
	if _, wasDeleted := s.vmctx.currentStateUpdate.Mutations.Dels[name]; wasDeleted {
		return nil, nil
	}
	ret, err := s.vmctx.stateDraft.Get(name)
	s.vmctx.GasBurn(gas.BurnCodeReadFromState1P, uint64(len(ret)/100)+1) // minimum 1
	return ret, err
}

func (s chainStateWrapper) Del(name kv.Key) {
	s.vmctx.currentStateUpdate.Mutations.Del(name)
}

func (s chainStateWrapper) Set(name kv.Key, value []byte) {
	s.vmctx.currentStateUpdate.Mutations.Set(name, value)
	// only burning gas when storing bytes to the state
	s.vmctx.GasBurn(gas.BurnCodeStorage1P, uint64(len(name)+len(value)))
}

func (vmctx *VMContext) State() kv.KVStore {
	return subrealm.New(vmctx.chainState(), kv.Key(vmctx.CurrentContractHname().Bytes()))
}

func (vmctx *VMContext) StateReader() kv.KVStoreReader {
	return subrealm.NewReadOnly(vmctx.chainState(), kv.Key(vmctx.CurrentContractHname().Bytes()))
}

// Disabled because of recursive calls
//func (vmctx *VMContext) State(burnGas ...kv.BurnGasFn) kv.KVStore {
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
