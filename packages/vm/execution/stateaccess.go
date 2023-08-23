package execution

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

type kvStoreWithGasBurn struct {
	kv.KVStore
	gas GasContext
}

func NewKVStoreWithGasBurn(s kv.KVStore, gas GasContext) kv.KVStore {
	return &kvStoreWithGasBurn{
		KVStore: s,
		gas:     gas,
	}
}

func (s *kvStoreWithGasBurn) Get(name kv.Key) []byte {
	return getWithGasBurn(s.KVStore, name, s.gas)
}

func (s *kvStoreWithGasBurn) Set(name kv.Key, value []byte) {
	s.KVStore.Set(name, value)
	s.gas.GasBurn(gas.BurnCodeStorage1P, uint64(len(name)+len(value)))
}

type kvStoreReaderWithGasBurn struct {
	kv.KVStoreReader
	gas GasContext
}

func NewKVStoreReaderWithGasBurn(r kv.KVStoreReader, gas GasContext) kv.KVStoreReader {
	return &kvStoreReaderWithGasBurn{
		KVStoreReader: r,
		gas:           gas,
	}
}

func (s *kvStoreReaderWithGasBurn) Get(name kv.Key) []byte {
	return getWithGasBurn(s.KVStoreReader, name, s.gas)
}

func getWithGasBurn(r kv.KVStoreReader, name kv.Key, gasctx GasContext) []byte {
	v := r.Get(name)
	// don't burn gas on GetNonce ("\xc1\x02\xcb\assn") or GetCode ("\xc1\x02\xcb\assc") calls
	// TODO: figure out a better way to skip EVM gas burn for these calls
	if !name.HasPrefix("\xc1\x02\xcb\assn") && !name.HasPrefix("\xc1\x02\xcb\assc") {
		gasctx.GasBurn(gas.BurnCodeReadFromState1P, uint64(len(v)/100)+1) // minimum 1
	}
	return v
}
