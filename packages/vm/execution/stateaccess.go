package execution

import (
	"fortio.org/safecast"
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
	// Safe conversion to avoid integer overflow
	size := len(name) + len(value)
	sizeUint := safecast.MustConvert[uint64](size)
	s.gas.GasBurn(gas.BurnCodeStorage1P, sizeUint)
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
	// Safe conversion to avoid integer overflow
	size := len(v) / 100
	sizeUint := safecast.MustConvert[uint64](size)
	gasctx.GasBurn(gas.BurnCodeReadFromState1P, sizeUint+1) // minimum 1
	return v
}
