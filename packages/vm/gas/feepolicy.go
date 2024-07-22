package gas

import (
	"fmt"
	"io"
	"math"
	"math/big"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

// By default each token pays for 100 units of gas
var DefaultGasPerToken = util.Ratio32{A: 100, B: 1}

// GasPerToken + ValidatorFeeShare + EVMGasRatio
const FeePolicyByteSize = util.RatioByteSize + serializer.OneByte + util.RatioByteSize

type FeePolicy struct {
	// EVMGasRatio expresses the ratio at which EVM gas is converted to ISC gas
	// X = ISC gas, Y = EVM gas => ISC gas = EVM gas * A/B
	EVMGasRatio util.Ratio32 `json:"evmGasRatio" swagger:"desc(The EVM gas ratio (ISC gas = EVM gas * A/B)),required"`

	// GasPerToken specifies how many gas units are paid for each token.
	GasPerToken util.Ratio32 `json:"gasPerToken" swagger:"desc(The gas per token ratio (A/B) (gas/token)),required"`

	// ValidatorFeeShare Validator/Governor fee split: percentage of fees which goes to Validator
	// 0 mean all goes to Governor
	// >=100 all goes to Validator
	ValidatorFeeShare uint8 `json:"validatorFeeShare" swagger:"desc(The validator fee share.),required"`
}

// FeeFromGasBurned calculates the how many tokens to take and where
// to deposit them.
// if gasPriceEVM == nil, the fee is calculated using the ISC GasPerToken
// price. Otherwise, the given gasPrice (expressed in base tokens with 'full
// decimals') is used instead.
func (p *FeePolicy) FeeFromGasBurned(gasUnits, availableTokens uint64, gasPrice *big.Int, l1BaseTokenDecimals uint32) (sendToOwner, sendToValidator uint64) {
	fee := p.FeeFromGas(gasUnits, gasPrice, l1BaseTokenDecimals)
	fee = min(fee, availableTokens)

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

// FeeFromGasWithGasPrice calculates the gas fee using the given gasPrice
// (expressed in ISC base tokens with 'full decimals').
func FeeFromGasWithGasPrice(gasUnits uint64, gasPrice *big.Int, l1BaseTokenDecimals uint32) uint64 {
	feeFullDecimals := new(big.Int).SetUint64(gasUnits)
	feeFullDecimals.Mul(feeFullDecimals, gasPrice)
	fee, remainder := util.EthereumDecimalsToBaseTokenDecimals(feeFullDecimals, l1BaseTokenDecimals)
	if remainder != nil && remainder.Sign() != 0 {
		fee++
	}
	return fee
}

// FeeFromGasWithGasPerToken calculates the gas fee using the ISC GasPerToken price
func FeeFromGasWithGasPerToken(gasUnits uint64, gasPerToken util.Ratio32) uint64 {
	if gasPerToken.IsEmpty() {
		return 0
	}
	return gasPerToken.YCeil64(gasUnits)
}

func (p *FeePolicy) FeeFromGas(gasUnits uint64, gasPrice *big.Int, l1BaseTokenDecimals uint32) uint64 {
	if p.GasPerToken.IsEmpty() {
		return 0
	}
	if gasPrice == nil {
		return FeeFromGasWithGasPerToken(gasUnits, p.GasPerToken)
	}
	return FeeFromGasWithGasPrice(gasUnits, gasPrice, l1BaseTokenDecimals)
}

func (p *FeePolicy) MinFee(gasPrice *big.Int, l1BaseTokenDecimals uint32) uint64 {
	return p.FeeFromGas(BurnCodeMinimumGasPerRequest1P.Cost(), gasPrice, l1BaseTokenDecimals)
}

func (p *FeePolicy) IsEnoughForMinimumFee(availableTokens uint64, gasPrice *big.Int, l1BaseTokenDecimals uint32) bool {
	return availableTokens >= p.MinFee(gasPrice, l1BaseTokenDecimals)
}

func (p *FeePolicy) GasBudgetFromTokens(availableTokens uint64, gasPrice *big.Int, l1BaseTokenDecimals uint32) uint64 {
	if p.GasPerToken.IsEmpty() {
		return math.MaxUint64
	}
	if gasPrice != nil {
		gasBudget := util.BaseTokensDecimalsToEthereumDecimals(availableTokens, l1BaseTokenDecimals)
		gasBudget.Div(gasBudget, gasPrice)
		return gasBudget.Uint64()
	}
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

func FeePolicyFromBytes(data []byte) (*FeePolicy, error) {
	return rwutil.ReadFromBytes(data, new(FeePolicy))
}

func (p *FeePolicy) Bytes() []byte {
	return rwutil.WriteToBytes(p)
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

func (p *FeePolicy) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	rr.Read(&p.EVMGasRatio)
	rr.Read(&p.GasPerToken)
	p.ValidatorFeeShare = rr.ReadUint8()
	return rr.Err
}

func (p *FeePolicy) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.Write(&p.EVMGasRatio)
	ww.Write(&p.GasPerToken)
	ww.WriteUint8(p.ValidatorFeeShare)
	return ww.Err
}

// DefaultGasPriceFullDecimals returns the default gas price to be set in EVM
// transactions, when using the ISC GasPerToken.
func (p *FeePolicy) DefaultGasPriceFullDecimals(l1BaseTokenDecimals uint32) *big.Int {
	// special case '0:0' for free request
	if p.GasPerToken.IsEmpty() {
		return big.NewInt(0)
	}

	// convert to wei (18 decimals)
	decimalsDifference := 18 - l1BaseTokenDecimals
	price := big.NewInt(10)
	price.Exp(price, new(big.Int).SetUint64(uint64(decimalsDifference)), nil)

	price.Mul(price, new(big.Int).SetUint64(uint64(p.GasPerToken.B)))
	price.Div(price, new(big.Int).SetUint64(uint64(p.GasPerToken.A)))
	price.Mul(price, new(big.Int).SetUint64(uint64(p.EVMGasRatio.A)))
	price.Div(price, new(big.Int).SetUint64(uint64(p.EVMGasRatio.B)))

	return price
}
