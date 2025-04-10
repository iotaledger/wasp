package webapi_validation

import (
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/tools/stardust-migration/webapi-validation/base"
)

type ChainValidation struct {
	base.ValidationContext
}

func NewChainValidation(validationContext base.ValidationContext) ChainValidation {
	return ChainValidation{
		ValidationContext: validationContext,
	}
}

func (c *ChainValidation) Validate(stateIndex uint32) {
	c.validateChainInfo(stateIndex)
}

func (c *ChainValidation) validateChainInfo(stateIndex uint32) {
	sRes, _, err := c.SClient.ChainsApi.GetChainInfo(c.Ctx, base.MainnetChainID).Execute()
	require.NoError(base.T, err)

	rRes, _, err := c.RClient.ChainsAPI.GetChainInfo(c.Ctx).Execute()
	require.NoError(base.T, err)

	require.Equal(base.T, sRes.EvmChainId, rRes.EvmChainId)
	require.Equal(base.T, sRes.IsActive, rRes.IsActive)

	// As we work with two different clients, we can not simply require.Equal(gasFeePolicy, gasFeePolicy) as the types are different, and require does do not a type independent reflection.
	// The types are different even with the same fields, leading to a failure.
	require.Equal(base.T, sRes.GasFeePolicy.EvmGasRatio.A, rRes.GasFeePolicy.EvmGasRatio.A)
	require.Equal(base.T, sRes.GasFeePolicy.EvmGasRatio.B, rRes.GasFeePolicy.EvmGasRatio.B)

	require.Equal(base.T, sRes.GasFeePolicy.GasPerToken.A, rRes.GasFeePolicy.GasPerToken.A)
	require.Equal(base.T, sRes.GasFeePolicy.GasPerToken.B, rRes.GasFeePolicy.GasPerToken.B)

	require.Equal(base.T, sRes.GasFeePolicy.ValidatorFeeShare, rRes.GasFeePolicy.ValidatorFeeShare)
}
