package governance

import (
	"github.com/iotaledger/hive.go/marshalutil"
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
	// Validator/Governor fee split: percentage of fees which goes to Validator
	// 0 mean all goes to Governor
	// >=100 all goes to Validator
	ValidatorFeeShare uint8
}

func DefaultGasFeePolicy() *GasFeePolicy {
	return &GasFeePolicy{
		GasFeeTokenID:     nil, // default is iotas
		FixedGasBudget:    nil,
		GasPerGasToken:    1,
		ValidatorFeeShare: 0, // by default all goes to the governor
	}
}

func MustGasFeePolicyFromBytes(data []byte) *GasFeePolicy {
	ret, err := GasFeePolicyFromBytes(data)
	if err != nil {
		panic(err)
	}
	return ret
}

func GasFeePolicyFromBytes(data []byte) (*GasFeePolicy, error) {
	ret := &GasFeePolicy{}
	mu := marshalutil.New(data)
	var gasNativeToken, fixedGasBudget bool
	var err error
	if gasNativeToken, err = mu.ReadBool(); err != nil {
		return nil, err
	}
	if gasNativeToken {
		b, err := mu.ReadBytes(iotago.NativeTokenIDLength)
		if err != nil {
			return nil, err
		}
		ret.GasFeeTokenID = &iotago.NativeTokenID{}
		copy(ret.GasFeeTokenID[:], b)
	}
	if fixedGasBudget, err = mu.ReadBool(); err != nil {
		return nil, err
	}
	if fixedGasBudget {
		budget, err := mu.ReadUint64()
		if err != nil {
			return nil, err
		}
		ret.FixedGasBudget = &budget
	}
	if ret.GasPerGasToken, err = mu.ReadUint64(); err != nil {
		return nil, err
	}
	if ret.ValidatorFeeShare, err = mu.ReadUint8(); err != nil {
		return nil, err
	}
	return ret, nil
}

func (g *GasFeePolicy) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteBool(g.GasFeeTokenID != nil)
	if g.GasFeeTokenID != nil {
		mu.WriteBytes(g.GasFeeTokenID[:])
	}
	mu.WriteBool(g.FixedGasBudget != nil)
	if g.FixedGasBudget != nil {
		mu.WriteUint64(*g.FixedGasBudget)
	}
	mu.WriteUint64(g.GasPerGasToken)
	mu.WriteUint8(g.ValidatorFeeShare)
	return mu.Bytes()
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

// GetGasFeePolicy returns gas policy from the state TODO
func GetGasFeePolicy(state kv.KVStoreReader) *GasFeePolicy {
	return MustGasFeePolicyFromBytes(state.MustGet(VarGasFeePolicyBytes))
}
