package util_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/util"
)

func TestBaseTokensDecimalsToEthereumDecimals(t *testing.T) {
	value := coin.Value(12345678)
	tests := []struct {
		decimals          uint8
		expected          uint64
		expectedRemainder uint64
	}{
		{
			decimals: 6,
			expected: 12345678000000000000,
		},
		{
			decimals: 18,
			expected: 12345678,
		},
	}
	for _, test := range tests {
		wei := util.BaseTokensDecimalsToEthereumDecimals(value, test.decimals)
		require.EqualValues(t, test.expected, wei.Uint64())
	}
}

func TestEthereumDecimalsToBaseTokenDecimals(t *testing.T) {
	value := uint64(123456789123456789)
	tests := []struct {
		decimals          uint8
		expected          uint64
		expectedRemainder uint64
	}{
		{
			decimals:          6,
			expected:          123456,
			expectedRemainder: 789123456789,
		},
		{
			decimals: 18,
			expected: value,
		},
	}
	for _, test := range tests {
		bt, rem := util.EthereumDecimalsToBaseTokenDecimals(new(big.Int).SetUint64(value), test.decimals)
		require.EqualValues(t, test.expected, bt)
		require.EqualValues(t, test.expectedRemainder, rem.Uint64())
	}
}
