package gas

import (
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/core/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/util"
)

var (
	emptyNativeTokenID = iotago.NativeTokenID{}

	// By default each token pays for 100 units of gas
	DefaultGasPerToken = util.Ratio32{A: 1, B: 100}
)

type GasFeePolicy struct {
	// GasFeeTokenID contains iotago.NativeTokenID used to pay for gas, or nil if base token are used for gas fee
	GasFeeTokenID iotago.NativeTokenID
	// GasFeeTokenDecimals the number of decimals in the native token used to pay for gas fees. Only considered if GasFeeTokenID != nil
	GasFeeTokenDecimals uint32

	// GasPerToken specifies how many gas units are paid for each token.
	GasPerToken util.Ratio32 // X = fee, Y = gas => fee = gas * A/B

	// EVMGasRatio expresses the ratio at which EVM gas is converted to ISC gas
	EVMGasRatio util.Ratio32 // X = ISC gas, Y = EVM gas => ISC gas = EVM gas * A/B

	// ValidatorFeeShare Validator/Governor fee split: percentage of fees which goes to Validator
	// 0 mean all goes to Governor
	// >=100 all goes to Validator
	ValidatorFeeShare uint8
}

// FeeFromGas calculates fee = gas * A/B
func FeeFromGas(gasUnits uint64, gasPerToken util.Ratio32) uint64 {
	return gasPerToken.XCeil64(gasUnits)
}

func (p *GasFeePolicy) FeeFromGas(gasUnits uint64) uint64 {
	return FeeFromGas(gasUnits, p.GasPerToken)
}

// FeeFromGasBurned calculates the how many tokens to take and where
// to deposit them.
func (p *GasFeePolicy) FeeFromGasBurned(gasUnits, availableTokens uint64) (sendToOwner, sendToValidator uint64) {
	var fee uint64

	// round up
	fee = p.FeeFromGas(gasUnits)
	fee = util.MinUint64(fee, availableTokens)

	validatorPercentage := p.ValidatorFeeShare
	if validatorPercentage > 100 {
		validatorPercentage = 100
	}
	// safe arithmetics
	if fee >= 100 {
		sendToValidator = (fee / 100) * uint64(validatorPercentage)
	} else {
		sendToValidator = (fee * uint64(validatorPercentage)) / 100
	}
	return fee - sendToValidator, sendToValidator
}

func (p *GasFeePolicy) MinFee() uint64 {
	return p.FeeFromGas(BurnCodeMinimumGasPerRequest1P.Cost())
}

func (p *GasFeePolicy) IsEnoughForMinimumFee(availableTokens uint64) bool {
	return availableTokens >= p.MinFee()
}

func (p *GasFeePolicy) GasBudgetFromTokens(availableTokens uint64) uint64 {
	return p.GasPerToken.YFloor64(availableTokens)
}

func DefaultGasFeePolicy() *GasFeePolicy {
	return &GasFeePolicy{
		GasFeeTokenID:     iotago.NativeTokenID{}, // default is base token
		GasPerToken:       DefaultGasPerToken,
		ValidatorFeeShare: 0, // by default all goes to the governor
		EVMGasRatio:       DefaultEVMGasRatio,
	}
}

func MustGasFeePolicyFromBytes(data []byte) *GasFeePolicy {
	ret, err := FeePolicyFromBytes(data)
	if err != nil {
		panic(err)
	}
	return ret
}

var ErrInvalidRatio = errors.New("ratio must have both components != 0")

func FeePolicyFromBytes(data []byte) (*GasFeePolicy, error) {
	return FeePolicyFromMarshalUtil(marshalutil.New(data))
}

func FeePolicyFromMarshalUtil(mu *marshalutil.MarshalUtil) (*GasFeePolicy, error) {
	ret := &GasFeePolicy{}
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
	if ret.GasPerToken, err = ReadRatio32(mu); err != nil {
		return nil, err
	}
	if ret.ValidatorFeeShare, err = mu.ReadUint8(); err != nil {
		return nil, err
	}
	if ret.EVMGasRatio, err = ReadRatio32(mu); err != nil {
		return nil, err
	}
	return ret, nil
}

func ReadRatio32(mu *marshalutil.MarshalUtil) (ret util.Ratio32, err error) {
	b, err := mu.ReadBytes(8)
	if err != nil {
		return ret, err
	}
	if ret, err = util.Ratio32FromBytes(b); err != nil {
		return ret, err
	}
	if ret.HasZeroComponent() {
		return ret, ErrInvalidRatio
	}
	return ret, nil
}

func (p *GasFeePolicy) Bytes() []byte {
	mu := marshalutil.New()
	hasGasFeeToken := p.GasFeeTokenID != emptyNativeTokenID
	mu.WriteBool(hasGasFeeToken)
	if hasGasFeeToken {
		mu.WriteBytes(p.GasFeeTokenID[:])
		mu.WriteUint32(p.GasFeeTokenDecimals)
	}
	mu.WriteBytes(p.GasPerToken.Bytes())
	mu.WriteUint8(p.ValidatorFeeShare)
	mu.WriteBytes(p.EVMGasRatio.Bytes())
	return mu.Bytes()
}

func (p *GasFeePolicy) String() string {
	return fmt.Sprintf(`
	GasFeeTokenID: %s
	GasFeeTokenDecimals %d
	GasPerToken %s
	EVMGasRatio %s
	ValidatorFeeShare %d
	`,
		p.GasFeeTokenID,
		p.GasFeeTokenDecimals,
		p.GasPerToken,
		p.EVMGasRatio,
		p.ValidatorFeeShare,
	)
}
