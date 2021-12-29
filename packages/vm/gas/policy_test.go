package gas_test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/vm/gas"

	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/stretchr/testify/require"
)

func TestFeePolicySerde(t *testing.T) {
	feePolicy := gas.DefaultGasFeePolicy()
	feePolicyBin := feePolicy.Bytes()
	feePolicyBack, err := gas.GasFeePolicyFromBytes(feePolicyBin)
	require.NoError(t, err)
	require.EqualValues(t, feePolicy.GasFeeTokenID, feePolicyBack.GasFeeTokenID)
	require.EqualValues(t, feePolicy.ValidatorFeeShare, feePolicyBack.ValidatorFeeShare)
	require.EqualValues(t, feePolicy.FixedGasBudget, feePolicyBack.FixedGasBudget)
	require.EqualValues(t, feePolicy.GasPricePerNominalUnit, feePolicyBack.GasPricePerNominalUnit)
	require.EqualValues(t, feePolicy.GasNominalUnit, feePolicyBack.GasNominalUnit)

	fgb := uint64(100)
	feePolicy = &gas.GasFeePolicy{
		GasFeeTokenID:          &tpkg.RandNativeToken().ID,
		FixedGasBudget:         nil,
		GasPricePerNominalUnit: fgb,
		ValidatorFeeShare:      10,
	}
	feePolicyBin = feePolicy.Bytes()
	feePolicyBack, err = gas.GasFeePolicyFromBytes(feePolicyBin)
	require.NoError(t, err)
	require.EqualValues(t, feePolicy.GasFeeTokenID, feePolicyBack.GasFeeTokenID)
	require.EqualValues(t, feePolicy.ValidatorFeeShare, feePolicyBack.ValidatorFeeShare)
	require.EqualValues(t, feePolicy.FixedGasBudget, feePolicyBack.FixedGasBudget)
	require.EqualValues(t, feePolicy.GasPricePerNominalUnit, feePolicyBack.GasPricePerNominalUnit)
	require.EqualValues(t, feePolicy.GasNominalUnit, feePolicyBack.GasNominalUnit)
}
