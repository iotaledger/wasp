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

// wei => custom token
func EthereumDecimalsToCustomTokenDecimals(value *big.Int, customTokenDecimals uint32) *big.Int {
	return adaptDecimals(value, ethereumDecimals, customTokenDecimals)
}

// custom token => wei
func CustomTokensDecimalsToEthereumDecimals(value *big.Int, customTokenDecimals uint32) *big.Int {
	return adaptDecimals(value, customTokenDecimals, ethereumDecimals)
}
