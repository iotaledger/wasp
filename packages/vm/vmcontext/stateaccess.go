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

func (s chainStateWrapper) Has(name kv.Key) bool {
	if _, ok := s.vmctx.currentStateUpdate.Mutations.Sets[name]; ok {
		return true
	}
	if _, wasDeleted := s.vmctx.currentStateUpdate.Mutations.Dels[name]; wasDeleted {
		return false
	}
	return s.vmctx.task.StateDraft.Has(name)
}

func (s chainStateWrapper) Iterate(prefix kv.Key, f func(kv.Key, []byte) bool) {
	s.IterateKeys(prefix, func(k kv.Key) bool {
		return f(k, s.Get(k))
	})
}

func (s chainStateWrapper) IterateKeys(prefix kv.Key, f func(key kv.Key) bool) {
	for k := range s.vmctx.currentStateUpdate.Mutations.Sets {
		if k.HasPrefix(prefix) {
			if !f(k) {
				return
			}
		}
	}
	s.vmctx.task.StateDraft.IterateKeys(prefix, func(k kv.Key) bool {
		if !s.vmctx.currentStateUpdate.Mutations.Contains(k) {
			return f(k)
		}
		return true
	})
}

func (s chainStateWrapper) IterateSorted(prefix kv.Key, f func(kv.Key, []byte) bool) {
	s.IterateKeysSorted(prefix, func(k kv.Key) bool {
		return f(k, s.Get(k))
	})
}

func (s chainStateWrapper) IterateKeysSorted(prefix kv.Key, f func(key kv.Key) bool) {
	var keys []kv.Key
	for k := range s.vmctx.currentStateUpdate.Mutations.Sets {
		if k.HasPrefix(prefix) {
			keys = append(keys, k)
		}
	}
	s.vmctx.task.StateDraft.IterateKeysSorted(prefix, func(k kv.Key) bool {
		if !s.vmctx.currentStateUpdate.Mutations.Contains(k) {
			keys = append(keys, k)
		}
		return true
	})
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	for _, k := range keys {
		if !f(k) {
			break
		}
	}
}

func (s chainStateWrapper) Get(name kv.Key) []byte {
	v, ok := s.vmctx.currentStateUpdate.Mutations.Sets[name]
	if ok {
		return v
	}
	if _, wasDeleted := s.vmctx.currentStateUpdate.Mutations.Dels[name]; wasDeleted {
		return nil
	}
	ret := s.vmctx.task.StateDraft.Get(name)
	s.vmctx.GasBurn(gas.BurnCodeReadFromState1P, uint64(len(ret)/100)+1) // minimum 1
	return ret
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

func (s chainStateWrapper) Apply() {
	s.vmctx.currentStateUpdate.Mutations.ApplyTo(s.vmctx.task.StateDraft)
}
