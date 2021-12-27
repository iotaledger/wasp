package testcore

import (
	"testing"

	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
)

func TestFeeBasic(t *testing.T) {
	env := solo.New(t)
	chain := env.NewChain(nil, "chain1")
	feePolicy := chain.GetGasFeePolicy()
	require.Nil(t, feePolicy.GasFeeTokenID)
	require.Nil(t, feePolicy.FixedGasBudget)
	require.EqualValues(t, 0, feePolicy.ValidatorFeeShare)
}
