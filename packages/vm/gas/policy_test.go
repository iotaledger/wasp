package gas

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/iota.go/v3/tpkg"
)

func TestFeePolicySerde(t *testing.T) {
	feePolicy := DefaultGasFeePolicy()
	feePolicyBin := feePolicy.Bytes()
	feePolicyBack, err := FeePolicyFromBytes(feePolicyBin)
	require.NoError(t, err)
	require.EqualValues(t, feePolicy.GasFeeTokenID, feePolicyBack.GasFeeTokenID)
	require.EqualValues(t, feePolicy.ValidatorFeeShare, feePolicyBack.ValidatorFeeShare)
	require.EqualValues(t, feePolicy.GasPerToken, feePolicyBack.GasPerToken)

	feePolicy = &GasFeePolicy{
		GasFeeTokenID:     tpkg.RandNativeToken().ID,
		GasPerToken:       DefaultGasPerToken,
		ValidatorFeeShare: 10,
		EVMGasRatio:       DefaultEVMGasRatio,
	}
	feePolicyBin = feePolicy.Bytes()
	feePolicyBack, err = FeePolicyFromBytes(feePolicyBin)
	require.NoError(t, err)
	require.EqualValues(t, feePolicy.GasFeeTokenID, feePolicyBack.GasFeeTokenID)
	require.EqualValues(t, feePolicy.ValidatorFeeShare, feePolicyBack.ValidatorFeeShare)
	require.EqualValues(t, feePolicy.GasPerToken, feePolicyBack.GasPerToken)
}
