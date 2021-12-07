package governance

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

const MinGasPerBlob = 1000

type GasFeePolicy struct {
	// GasFeeTokenID contains iotago.NativeTokenID used to pay for gas, or nil if iotas are used for gas fee
	GasFeeTokenID *iotago.NativeTokenID
	// FixedGasBudget != nil if assume each call has a fixed budget
	FixedGasBudget *uint64
	// GasPerGasToken how many gas you get for 1 iota or another gas token
	GasPerGasToken uint64
}

// GetGasFeePolicy returns gas policy from the state TODO
func GetGasFeePolicy(state kv.KVStoreReader) *GasFeePolicy {
	return &GasFeePolicy{
		GasPerGasToken: 1,
	}
}

func GasForBlob(blob dict.Dict) uint64 {
	gas := uint64(0)
	for k, v := range blob {
		gas += uint64(len(k)) + uint64(len(v))
	}
	if gas < MinGasPerBlob {
		gas = MinGasPerBlob
	}
	return gas
}
