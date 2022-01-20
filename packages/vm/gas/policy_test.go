package gas

import (
	"testing"

	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/stretchr/testify/require"
)

func TestFeePolicySerde(t *testing.T) {
	feePolicy := DefaultGasFeePolicy()
	feePolicyBin := feePolicy.Bytes()
	feePolicyBack, err := GasFeePolicyFromBytes(feePolicyBin)
	require.NoError(t, err)
	require.EqualValues(t, feePolicy.GasFeeTokenID, feePolicyBack.GasFeeTokenID)
	require.EqualValues(t, feePolicy.ValidatorFeeShare, feePolicyBack.ValidatorFeeShare)
	require.EqualValues(t, feePolicy.GasPricePerNominalUnit, feePolicyBack.GasPricePerNominalUnit)
	require.EqualValues(t, feePolicy.GasNominalUnit, feePolicyBack.GasNominalUnit)

	fgb := uint64(100)
	feePolicy = &GasFeePolicy{
		GasFeeTokenID:          &tpkg.RandNativeToken().ID,
		GasPricePerNominalUnit: fgb,
		ValidatorFeeShare:      10,
	}
	feePolicyBin = feePolicy.Bytes()
	feePolicyBack, err = GasFeePolicyFromBytes(feePolicyBin)
	require.NoError(t, err)
	require.EqualValues(t, feePolicy.GasFeeTokenID, feePolicyBack.GasFeeTokenID)
	require.EqualValues(t, feePolicy.ValidatorFeeShare, feePolicyBack.ValidatorFeeShare)
	require.EqualValues(t, feePolicy.GasPricePerNominalUnit, feePolicyBack.GasPricePerNominalUnit)
	require.EqualValues(t, feePolicy.GasNominalUnit, feePolicyBack.GasNominalUnit)
}
