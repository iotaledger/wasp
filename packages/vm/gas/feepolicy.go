package gas

import (
	"fmt"
	"math"
	"math/big"

	bcs "github.com/iotaledger/bcs-go"

	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/util"
)

// DefaultGasPerToken declares that each token pays for 10 units of gas
var DefaultGasPerToken = util.Ratio32{A: 1, B: 10}

type FeePolicy struct {
	// EVMGasRatio expresses the ratio at which EVM gas is converted to ISC gas
	// X = ISC gas, Y = EVM gas => ISC gas = EVM gas * A/B
	EVMGasRatio util.Ratio32 `json:"evmGasRatio" swagger:"desc(The EVM gas ratio (ISC gas = EVM gas * A/B)),required"`

	// GasPerToken specifies how many gas units are paid for each token.
	GasPerToken util.Ratio32 `json:"gasPerToken" swagger:"desc(The gas per token ratio (A/B) (gas/token)),required"`

	// ValidatorFeeShare is the Validator/ChainAdmin fee split: percentage of fees which goes to Validator
	// 0 means all goes to ChainAdmin
	// >=100 all goes to Validator
	ValidatorFeeShare uint8 `json:"validatorFeeShare" swagger:"desc(The validator fee share.),required"`
}

// FeeFromGasBurned calculates the how many tokens to take and where
// to deposit them.
// if gasPriceEVM == nil, the fee is calculated using the ISC GasPerToken
// price. Otherwise, the given evmGasPrice (expressed in base tokens with 'full
// decimals') is used instead.
func (p *FeePolicy) FeeFromGasBurned(gasUnits uint64, availableTokens coin.Value, evmGasPrice *big.Int, l1BaseTokenDecimals uint8) (sendToOwner, sendToValidator coin.Value) {
	fee := p.FeeFromGas(gasUnits, evmGasPrice, l1BaseTokenDecimals)
	fee = min(fee, availableTokens)

	validatorPercentage := min(100, coin.Value(p.ValidatorFeeShare))

	if fee >= 100 {
		sendToValidator = (fee / 100) * validatorPercentage
	} else {
		sendToValidator = (fee * validatorPercentage) / 100
	}

	sendToOwner = fee - sendToValidator
	return
}

// FeeFromGasWithGasPrice calculates the gas fee using the given evmGasPrice
// (expressed in ISC base tokens with 'full decimals').
func FeeFromGasWithGasPrice(gasUnits uint64, evmGasPrice *big.Int, l1BaseTokenDecimals uint8) coin.Value {
	feeFullDecimals := new(big.Int).Set(evmGasPrice)
	feeFullDecimals.Mul(feeFullDecimals, big.NewInt(0).SetUint64(gasUnits))
	fee, remainder := util.EthereumDecimalsToBaseTokenDecimals(feeFullDecimals, l1BaseTokenDecimals)
	if remainder != nil && remainder.Sign() != 0 {
		fee++
	}
	return fee
}

// FeeFromGasWithGasPerToken calculates the gas fee using the ISC GasPerToken price
func FeeFromGasWithGasPerToken(gasUnits uint64, gasPerToken util.Ratio32) coin.Value {
	if gasPerToken.IsEmpty() {
		return 0
	}
	return coin.Value(gasPerToken.YCeil64(gasUnits))
}

func (p *FeePolicy) FeeFromGas(gasUnits uint64, evmGasPrice *big.Int, l1BaseTokenDecimals uint8) coin.Value {
	if p.GasPerToken.IsEmpty() {
		return 0
	}
	if evmGasPrice == nil {
		return FeeFromGasWithGasPerToken(gasUnits, p.GasPerToken)
	}
	return FeeFromGasWithGasPrice(gasUnits, evmGasPrice, l1BaseTokenDecimals)
}

func (p *FeePolicy) MinFee(evmGasPrice *big.Int, l1BaseTokenDecimals uint8) coin.Value {
	return p.FeeFromGas(BurnCodeMinimumGasPerRequest1P.Cost(), evmGasPrice, l1BaseTokenDecimals)
}

func (p *FeePolicy) IsEnoughForMinimumFee(availableTokens coin.Value, evmGasPrice *big.Int, l1BaseTokenDecimals uint8) bool {
	// return availableTokens >= p.MinFee(evmGasPrice, l1BaseTokenDecimals)
	return availableTokens >= p.MinFee(evmGasPrice, l1BaseTokenDecimals)
}

func (p *FeePolicy) GasBudgetFromTokensFullDecimals(availableTokens *big.Int, evmGasPrice *big.Int) uint64 {
	if p.GasPerToken.IsEmpty() {
		return math.MaxUint64
	}
	gasBudget := new(big.Int).Set(availableTokens)
	gasBudget.Div(gasBudget, evmGasPrice)
	return gasBudget.Uint64()
}

func (p *FeePolicy) GasBudgetFromTokensWithGasPrice(availableTokens coin.Value, evmGasPrice *big.Int, decimals uint8) uint64 {
	return p.GasBudgetFromTokensFullDecimals(
		util.BaseTokensDecimalsToEthereumDecimals(availableTokens, decimals),
		evmGasPrice,
	)
}

func (p *FeePolicy) GasBudgetFromTokens(availableTokens coin.Value) uint64 {
	if p.GasPerToken.IsEmpty() {
		return math.MaxUint64
	}
	return p.GasPerToken.XFloor64(uint64(availableTokens))
}

func DefaultFeePolicy() *FeePolicy {
	return &FeePolicy{
		GasPerToken:       DefaultGasPerToken,
		ValidatorFeeShare: 0, // by default all goes to the chain admin
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
	return bcs.Unmarshal[*FeePolicy](data)
}

func (p *FeePolicy) Bytes() []byte {
	return bcs.MustMarshal(p)
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

// DefaultGasPriceFullDecimals returns the default gas price to be set in EVM
// transactions, when using the ISC GasPerToken.
func (p *FeePolicy) DefaultGasPriceFullDecimals(l1BaseTokenDecimals uint8) *big.Int {
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
