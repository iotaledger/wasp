package util

import "math/big"

const ethereumDecimals = uint32(18)

func adaptDecimals(value *big.Int, fromDecimals, toDecimals uint32) *big.Int {
	v := new(big.Int).Set(value) // clone value
	exp := big.NewInt(10)
	if toDecimals > fromDecimals {
		exp.Exp(exp, big.NewInt(int64(toDecimals-fromDecimals)), nil)
		return v.Mul(v, exp)
	}
	exp.Exp(exp, big.NewInt(int64(fromDecimals-toDecimals)), nil)
	return v.Div(v, exp)
}

// wei => base tokens
func EthereumDecimalsToBaseTokenDecimals(value *big.Int, baseTokenDecimals uint32) uint64 {
	v := adaptDecimals(value, ethereumDecimals, baseTokenDecimals)
	if !v.IsUint64() {
		panic("cannot convert ether value to base tokens: too large")
	}
	return v.Uint64()
}

// base tokens => wei
func BaseTokensDecimalsToEthereumDecimals(value uint64, baseTokenDecimals uint32) *big.Int {
	return adaptDecimals(new(big.Int).SetUint64(value), baseTokenDecimals, ethereumDecimals)
}
