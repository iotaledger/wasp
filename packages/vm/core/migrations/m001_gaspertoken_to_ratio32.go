package migrations

import (
	"math"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/hive.go/core/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

var m001GasPerTokenToRatio32 = Migration{
	Contract: governance.Contract,

	Apply: func(state kv.KVStore, log *logger.Logger) error {
		fpOld, err := feePolicyFromBytesOld(state.MustGet(governance.VarGasFeePolicyBytes))
		if err != nil {
			return err
		}
		if fpOld.GasPerToken > math.MaxUint32 {
			log.Warn("m001GasPerTokenToRatio32: trimming gas per token")
			fpOld.GasPerToken = math.MaxUint32
		}
		fpNew := &gas.GasFeePolicy{
			GasFeeTokenID:       fpOld.GasFeeTokenID,
			GasFeeTokenDecimals: fpOld.GasFeeTokenDecimals,
			GasPerToken:         util.Ratio32{A: 1, B: uint32(fpOld.GasPerToken)},
			EVMGasRatio:         fpOld.EVMGasRatio,
			ValidatorFeeShare:   fpOld.ValidatorFeeShare,
		}
		state.Set(governance.VarGasFeePolicyBytes, fpNew.Bytes())
		return nil
	},
}

type gasFeePolicyOld struct {
	GasFeeTokenID       iotago.NativeTokenID
	GasFeeTokenDecimals uint32
	GasPerToken         uint64
	EVMGasRatio         util.Ratio32
	ValidatorFeeShare   uint8
}

func feePolicyFromBytesOld(data []byte) (*gasFeePolicyOld, error) {
	ret := &gasFeePolicyOld{}
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
	if ret.GasPerToken, err = mu.ReadUint64(); err != nil {
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
