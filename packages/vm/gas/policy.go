package gas

import (
	"math"

	"github.com/iotaledger/hive.go/core/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/util"
)

type GasFeePolicy struct {
	// GasFeeTokenID contains iotago.NativeTokenID used to pay for gas, or nil if base token are used for gas fee
	GasFeeTokenID *iotago.NativeTokenID
	// GasPerToken specifies how many gas units are paid for each token ( 100 means 1 tokens pays for 100 gas)
	GasPerToken uint64
	// ValidatorFeeShare Validator/Governor fee split: percentage of fees which goes to Validator
	// 0 mean all goes to Governor
	// >=100 all goes to Validator
	ValidatorFeeShare uint8
}

func calcFee(gasUnits, gasPerToken uint64) uint64 {
	return uint64(math.Ceil(float64(gasUnits) / float64(gasPerToken)))
}

// FeeFromGas return ownerFee and validatorFee
func (p *GasFeePolicy) FeeFromGas(gasUnits, availableTokens uint64) (sendToOwner, sendToValidator uint64) {
	var fee uint64

	// round up
	fee = calcFee(gasUnits, p.GasPerToken)
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
	return calcFee(BurnCodeMinimumGasPerRequest1P.Cost(), p.GasPerToken)
}

func (p *GasFeePolicy) IsEnoughForMinimumFee(availableTokens uint64) bool {
	return availableTokens >= p.MinFee()
}

func (p *GasFeePolicy) AffordableGasBudgetFromAvailableTokens(availableTokens uint64) uint64 {
	return availableTokens * p.GasPerToken
}

func DefaultGasFeePolicy() *GasFeePolicy {
	return &GasFeePolicy{
		GasFeeTokenID:     nil, // default is base token
		GasPerToken:       100, // each token pays for 100 units of gas
		ValidatorFeeShare: 0,   // by default all goes to the governor
	}
}

func MustGasFeePolicyFromBytes(data []byte) *GasFeePolicy {
	ret, err := FeePolicyFromBytes(data)
	if err != nil {
		panic(err)
	}
	return ret
}

func FeePolicyFromBytes(data []byte) (*GasFeePolicy, error) {
	ret := &GasFeePolicy{}
	mu := marshalutil.New(data)
	var gasNativeToken bool
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
	if ret.GasPerToken, err = mu.ReadUint64(); err != nil {
		return nil, err
	}
	if ret.ValidatorFeeShare, err = mu.ReadUint8(); err != nil {
		return nil, err
	}
	return ret, nil
}

func (p *GasFeePolicy) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteBool(p.GasFeeTokenID != nil)
	if p.GasFeeTokenID != nil {
		mu.WriteBytes(p.GasFeeTokenID[:])
	}
	mu.WriteUint64(p.GasPerToken)
	mu.WriteUint8(p.ValidatorFeeShare)
	return mu.Bytes()
}
