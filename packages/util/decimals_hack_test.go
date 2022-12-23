package util

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBaseTokensDecimalsToEthereumDecimals(t *testing.T) {
	value := big.NewInt(12345678)
	tests := []struct {
		decimals uint32
		expected string
	}{
		{
			decimals: 6,
			expected: "12345678000000000000",
		},
		{
			decimals: 18,
			expected: "12345678",
		},
		{
			decimals: 20,
			expected: "123456",
		},
	}
	for _, test := range tests {
		require.EqualValues(t,
			test.expected,
			CustomTokensDecimalsToEthereumDecimals(value, test.decimals).String(),
		)
	}
}

func TestEthereumDecimalsToBaseTokenDecimals(t *testing.T) {
	value := big.NewInt(123456789123456789)
	tests := []struct {
		decimals uint32
		expected string
	}{
		{
			decimals: 6,
			expected: "123456", // extra decimal cases will be ignored
		},
		{
			decimals: 18,
			expected: value.String(),
		},
		{
			decimals: 20,
			expected: value.String() + "00",
		},
	}
	for _, test := range tests {
		require.EqualValues(t,
			test.expected,
			EthereumDecimalsToCustomTokenDecimals(value, test.decimals).String(),
		)
	}
}
