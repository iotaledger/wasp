package vmcontext

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
)

func (vmctx *VMContext) addContractSubPartition(key kv.Key) kv.Key {
	return kv.Key(vmctx.ContractHname().Bytes()) + key
}

func (vmctx *VMContext) Has(name kv.Key) (bool, error) {
	name = vmctx.addContractSubPartition(name)
	mut := vmctx.stateUpdate.Mutations().Latest(name)
	if mut != nil {
		return mut.Value() != nil, nil
	}
	return vmctx.virtualState.Variables().Has(name)
}

func (vmctx *VMContext) Iterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) error {
	prefix = vmctx.addContractSubPartition(prefix)
	// TODO is ot correct?
	seen, done := vmctx.stateUpdate.Mutations().IterateValues(prefix, f)
	if done {
		return nil
	}
	return vmctx.virtualState.Variables().Iterate(prefix, func(key kv.Key, value []byte) bool {
		_, ok := seen[key]
		if ok {
			return true
		}
		return f(key, value)
	})
}

func (vmctx *VMContext) IterateKeys(prefix kv.Key, f func(key kv.Key) bool) error {
	prefix = vmctx.addContractSubPartition(prefix)
	seen, done := vmctx.stateUpdate.Mutations().IterateValues(prefix, func(key kv.Key, value []byte) bool {
		return f(key)
	})
	if done {
		return nil
	}
	return vmctx.virtualState.Variables().IterateKeys(prefix, func(key kv.Key) bool {
		_, ok := seen[key]
		if ok {
			return true
		}
		return f(key)
	})
}

func (vmctx *VMContext) Get(name kv.Key) ([]byte, error) {
	name = vmctx.addContractSubPartition(name)
	mut := vmctx.stateUpdate.Mutations().Latest(name)
	if mut != nil {
		return mut.Value(), nil
	}
	return vmctx.virtualState.Variables().Get(name)
}

func (vmctx *VMContext) Del(name kv.Key) {
	name = vmctx.addContractSubPartition(name)
	vmctx.stateUpdate.Mutations().Add(buffered.NewMutationDel(name))
}

func (vmctx *VMContext) Set(name kv.Key, value []byte) {
	name = vmctx.addContractSubPartition(name)
	vmctx.stateUpdate.Mutations().Add(buffered.NewMutationSet(name, value))
}
