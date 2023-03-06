package gas

import (
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
	"github.com/iotaledger/wasp/packages/util"
)

// By default each token pays for 100 units of gas
var DefaultGasPerToken = util.Ratio32{A: 100, B: 1}

type FeePolicy struct {
	// GasPerToken specifies how many gas units are paid for each token.
	GasPerToken util.Ratio32

	// EVMGasRatio expresses the ratio at which EVM gas is converted to ISC gas
	EVMGasRatio util.Ratio32 // X = ISC gas, Y = EVM gas => ISC gas = EVM gas * A/B

	// ValidatorFeeShare Validator/Governor fee split: percentage of fees which goes to Validator
	// 0 mean all goes to Governor
	// >=100 all goes to Validator
	ValidatorFeeShare uint8
}

// FeeFromGasBurned calculates the how many tokens to take and where
// to deposit them.
func (p *FeePolicy) FeeFromGasBurned(gasUnits, availableTokens uint64) (sendToOwner, sendToValidator uint64) {
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

func FeeFromGas(gasUnits uint64, gasPerToken util.Ratio32) uint64 {
	return gasPerToken.YCeil64(gasUnits)
}

func (p *FeePolicy) FeeFromGas(gasUnits uint64) uint64 {
	return FeeFromGas(gasUnits, p.GasPerToken)
}

func (p *FeePolicy) MinFee() uint64 {
	return p.FeeFromGas(BurnCodeMinimumGasPerRequest1P.Cost())
}

func (p *FeePolicy) IsEnoughForMinimumFee(availableTokens uint64) bool {
	return availableTokens >= p.MinFee()
}

func (p *FeePolicy) GasBudgetFromTokens(availableTokens uint64) uint64 {
	return p.GasPerToken.XFloor64(availableTokens)
}

func DefaultFeePolicy() *FeePolicy {
	return &FeePolicy{
		GasPerToken:       DefaultGasPerToken,
		ValidatorFeeShare: 0, // by default all goes to the governor
		EVMGasRatio:       DefaultEVMGasRatio,
	}
}

func MustFeePolicyFromBytes(data []byte) *FeePolicy {
	ret, err := FeePolicyFromBytes(data)
	if err != nil {
		panic(err)
	}
	return ret
}

var ErrInvalidRatio = errors.New("ratio must have both components != 0")

func FeePolicyFromBytes(data []byte) (*FeePolicy, error) {
	return FeePolicyFromMarshalUtil(marshalutil.New(data))
}

func FeePolicyFromMarshalUtil(mu *marshalutil.MarshalUtil) (*FeePolicy, error) {
	ret := &FeePolicy{}
	var err error
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

func (p *FeePolicy) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteBytes(p.GasPerToken.Bytes())
	mu.WriteUint8(p.ValidatorFeeShare)
	mu.WriteBytes(p.EVMGasRatio.Bytes())
	return mu.Bytes()
}

func (p *FeePolicy) String() string {
	return fmt.Sprintf(`
	GasPerToken %s
	EVMGasRatio %s
	ValidatorFeeShare %d
	`,
		p.GasPerToken,
		p.EVMGasRatio,
		p.ValidatorFeeShare,
	)
}
