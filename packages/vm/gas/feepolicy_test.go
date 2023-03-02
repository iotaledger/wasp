package gas

import (
	"testing"

	"github.com/stretchr/testify/require"
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
