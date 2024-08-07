package util

import (
	"math/big"

	"github.com/iotaledger/wasp/packages/coin"
)

const ethereumDecimals = uint32(18)

func adaptDecimals(value *big.Int, fromDecimals, toDecimals uint32) (result *big.Int, remainder *big.Int) {
	result = new(big.Int)
	remainder = new(big.Int)
	exp := big.NewInt(10)
	if toDecimals > fromDecimals {
		exp.Exp(exp, big.NewInt(int64(toDecimals-fromDecimals)), nil)
		result.Mul(value, exp)
	} else {
		exp.Exp(exp, big.NewInt(int64(fromDecimals-toDecimals)), nil)
		result.DivMod(value, exp, remainder)
	}
	return
}

// wei => base tokens
func EthereumDecimalsToBaseTokenDecimals(value *big.Int, baseTokenDecimals uint32) (result coin.Value, remainder *big.Int) {
	if baseTokenDecimals > ethereumDecimals {
		panic("expected baseTokenDecimals <= ethereumDecimals")
	}
	r, m := adaptDecimals(value, ethereumDecimals, baseTokenDecimals)
	if !r.IsUint64() {
		panic("cannot convert ether value to base tokens: too large")
	}
	return coin.Value(r.Uint64()), m
}

func MustEthereumDecimalsToBaseTokenDecimalsExact(value *big.Int, baseTokenDecimals uint32) (result coin.Value) {
	r, m := EthereumDecimalsToBaseTokenDecimals(value, baseTokenDecimals)
	if m.Sign() != 0 {
		panic("cannot convert ether value to base tokens: non-exact conversion")
	}
	return r
}

// base tokens => wei
func BaseTokensDecimalsToEthereumDecimals(value coin.Value, baseTokenDecimals uint32) (result *big.Int) {
	if baseTokenDecimals > ethereumDecimals {
		panic("expected baseTokenDecimals <= ethereumDecimals")
	}
	r, m := adaptDecimals(new(big.Int).SetUint64(uint64(value)), baseTokenDecimals, ethereumDecimals)
	if m.Sign() != 0 {
		panic("expected zero remainder")
	}
	return r
}
