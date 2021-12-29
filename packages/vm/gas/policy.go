package gas

import (
	"math"

	"github.com/iotaledger/wasp/packages/util"

	"github.com/iotaledger/hive.go/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
)

type GasFeePolicy struct {
	// GasFeeTokenID contains iotago.NativeTokenID used to pay for gas, or nil if iotas are used for gas fee
	GasFeeTokenID *iotago.NativeTokenID
	// FixedGasBudget != nil if assume each call has a fixed budget
	FixedGasBudget *uint64
	// GasNominalUnit price is specified per how many gas units
	GasNominalUnit uint64
	// GasPricePer1000Units how many gas tokens you pay for 1000 gas
	GasPricePerNominalUnit uint64
	// ValidatorFeeShare Validator/Governor fee split: percentage of fees which goes to Validator
	// 0 mean all goes to Governor
	// >=100 all goes to Validator
	ValidatorFeeShare uint8
}

// FeeFromGas return ownerFee and validatorFee
func (p *GasFeePolicy) FeeFromGas(g uint64, availableTokens ...uint64) (uint64, uint64) {
	available := uint64(math.MaxUint64)
	if len(availableTokens) > 0 {
		available = availableTokens[0]
	}
	var totalFee uint64
	nominalUnits := g / p.GasNominalUnit
	if nominalUnits == 0 {
		totalFee = util.MinUint64(available, p.GasPricePerNominalUnit)
	}
	if nominalUnits > math.MaxUint64/p.GasPricePerNominalUnit {
		totalFee = math.MaxUint64
	} else {
		totalFee = nominalUnits * p.GasPricePerNominalUnit
	}
	totalFee = util.MinUint64(totalFee, available)

	validatorPercentage := p.ValidatorFeeShare
	if validatorPercentage > 100 {
		validatorPercentage = 100
	}
	var sendToValidator uint64
	// safe arithmetics
	if totalFee >= 100 {
		sendToValidator = (totalFee / 100) * uint64(validatorPercentage)
	} else {
		sendToValidator = (totalFee * uint64(validatorPercentage)) / 100
	}
	return totalFee - sendToValidator, sendToValidator
}

func DefaultGasFeePolicy() *GasFeePolicy {
	return &GasFeePolicy{
		GasFeeTokenID:          nil, // default is iotas
		FixedGasBudget:         nil, // default is dynamic gas budget
		GasNominalUnit:         100, // gas is burned in 100-s and not less than 100
		GasPricePerNominalUnit: 100, // default is 1 gas == 1 iota
		ValidatorFeeShare:      0,   // by default all goes to the governor
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
	if ret.GasNominalUnit, err = mu.ReadUint64(); err != nil {
		return nil, err
	}
	if ret.GasPricePerNominalUnit, err = mu.ReadUint64(); err != nil {
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
	mu.WriteUint64(g.GasNominalUnit)
	mu.WriteUint64(g.GasPricePerNominalUnit)
	mu.WriteUint8(g.ValidatorFeeShare)
	return mu.Bytes()
}
