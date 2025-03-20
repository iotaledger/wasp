package gas

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util"
)

func TestFeePolicySerde(t *testing.T) {
	feePolicy := DefaultFeePolicy()
	feePolicyBin := feePolicy.Bytes()
	feePolicyBack, err := FeePolicyFromBytes(feePolicyBin)
	require.NoError(t, err)
	require.EqualValues(t, feePolicy.ValidatorFeeShare, feePolicyBack.ValidatorFeeShare)
	require.EqualValues(t, feePolicy.GasPerToken, feePolicyBack.GasPerToken)

	feePolicy = &FeePolicy{
		GasPerToken:       DefaultGasPerToken,
		ValidatorFeeShare: 10,
		EVMGasRatio:       DefaultEVMGasRatio,
	}
	feePolicyBin = feePolicy.Bytes()
	feePolicyBack, err = FeePolicyFromBytes(feePolicyBin)
	require.NoError(t, err)
	require.EqualValues(t, feePolicy.ValidatorFeeShare, feePolicyBack.ValidatorFeeShare)
	require.EqualValues(t, feePolicy.GasPerToken, feePolicyBack.GasPerToken)
}

func TestFeePolicyAffordableGas(t *testing.T) {
	feePolicy := DefaultFeePolicy()
	// needs 110 tokens for 1 gas
	feePolicy.GasPerToken = util.Ratio32{A: 1, B: 110}

	// map of [n tokens] expected gas
	cases := map[coin.Value]int{
		109: 0,
		200: 1,
		219: 1,
		220: 2,
	}
	for tokens, expectedGas := range cases {
		require.EqualValues(t, expectedGas, feePolicy.GasBudgetFromTokens(tokens))
	}

	// tokens charged for max gas
	// map of [n tokens] expected tokens charged
	cases2 := map[coin.Value]coin.Value{
		109: 0,
		110: 110,
		111: 110,
	}
	for tokens, expected := range cases2 {
		require.EqualValues(t, expected, feePolicy.FeeFromGas(
			feePolicy.GasBudgetFromTokens(tokens),
			nil, parameters.BaseTokenDecimals,
		))
	}
}
