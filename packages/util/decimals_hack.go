package util

import "math/big"

const ethereumDecimals = 18

func adaptDecimals(value *big.Int, fromDecimals, toDecimals int64) *big.Int {
	v := new(big.Int).Set(value) // clone value
	exp := big.NewInt(10)
	if toDecimals > fromDecimals {
		exp.Exp(exp, big.NewInt(toDecimals-fromDecimals), nil)
		return v.Mul(v, exp)
	}
	exp.Exp(exp, big.NewInt(fromDecimals-toDecimals), nil)
	return v.Div(v, exp)
}

// wei => base token
func EthereumDecimalsToBaseTokenDecimals(value *big.Int, baseTokenDecimals int64) *big.Int {
	return adaptDecimals(value, ethereumDecimals, baseTokenDecimals)
}

// base token => wei
func BaseTokensDecimalsToEthereumDecimals(value *big.Int, baseTokenDecimals int64) *big.Int {
	return adaptDecimals(value, baseTokenDecimals, ethereumDecimals)
}
