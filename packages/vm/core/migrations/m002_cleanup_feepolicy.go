package migrations

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

var m002CleanupFeePolicy = Migration{
	Contract: governance.Contract,

	Apply: func(state kv.KVStore, log *logger.Logger) error {
		fpBinOld := state.MustGet(governance.VarGasFeePolicyBytes)
		fpNew, err := m002ConvertFeePolicy(fpBinOld)
		if err != nil {
			return err
		}
		state.Set(governance.VarGasFeePolicyBytes, fpNew.Bytes())
		return nil
	},
}

func m002ConvertFeePolicy(oldBin []byte) (*gas.FeePolicy, error) {
	fpOld, err := feePolicyPreCleanupFromBytes(oldBin)
	if err != nil {
		return nil, err
	}
	return &gas.FeePolicy{
		GasPerToken:       fpOld.GasPerToken,
		EVMGasRatio:       fpOld.EVMGasRatio,
		ValidatorFeeShare: fpOld.ValidatorFeeShare,
	}, nil
}

type feePolicyPreCleanup struct {
	GasFeeTokenID       iotago.NativeTokenID
	GasFeeTokenDecimals uint32
	GasPerToken         util.Ratio32
	EVMGasRatio         util.Ratio32
	ValidatorFeeShare   uint8
}

func feePolicyPreCleanupFromBytes(data []byte) (*feePolicyPreCleanup, error) {
	ret := &feePolicyPreCleanup{}
	mu := marshalutil.New(data)
	var gasNativeToken bool
	var err error
	if gasNativeToken, err = mu.ReadBool(); err != nil {
		return nil, err
	}
	if gasNativeToken {
		b, err2 := mu.ReadBytes(iotago.NativeTokenIDLength)
		if err2 != nil {
			return nil, err2
		}
		ret.GasFeeTokenID = iotago.NativeTokenID{}
		copy(ret.GasFeeTokenID[:], b)
		if ret.GasFeeTokenDecimals, err2 = mu.ReadUint32(); err2 != nil {
			return nil, err2
		}
	}
	if ret.GasPerToken, err = gas.ReadRatio32(mu); err != nil {
		return nil, err
	}
	if ret.ValidatorFeeShare, err = mu.ReadUint8(); err != nil {
		return nil, err
	}
	if ret.EVMGasRatio, err = gas.ReadRatio32(mu); err != nil {
		return nil, err
	}
	return ret, nil
}
