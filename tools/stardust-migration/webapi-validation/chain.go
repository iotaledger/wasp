package webapi_validation

import (
	"github.com/stretchr/testify/require"
)

type ChainValidation struct {
	ValidationContext
}

func NewChainValidation(validationContext ValidationContext) ChainValidation {
	return ChainValidation{
		ValidationContext: validationContext,
	}
}

func (c *ChainValidation) Validate(stateIndex uint32) {
	c.validateChainInfo(stateIndex)
}

func (c *ChainValidation) validateChainInfo(stateIndex uint32) {
	sRes, _, err := c.sClient.ChainsApi.GetChainInfo(c.ctx, MainnetChainID).Execute()
	require.NoError(t, err)

	rRes, _, err := c.rClient.ChainsAPI.GetChainInfo(c.ctx).Execute()
	require.NoError(t, err)

	require.Equal(t, sRes.EvmChainId, rRes.EvmChainId)
	require.Equal(t, sRes.IsActive, rRes.IsActive)

	// As we work with two different clients, we can not simply require.Equal(gasFeePolicy, gasFeePolicy) as the types are different, and require does do not a type independent reflection.
	// The types are different even with the same fields, leading to a failure.
	require.Equal(t, sRes.GasFeePolicy.EvmGasRatio.A, rRes.GasFeePolicy.EvmGasRatio.A)
	require.Equal(t, sRes.GasFeePolicy.EvmGasRatio.B, rRes.GasFeePolicy.EvmGasRatio.B)

	require.Equal(t, sRes.GasFeePolicy.GasPerToken.A, rRes.GasFeePolicy.GasPerToken.A)
	require.Equal(t, sRes.GasFeePolicy.GasPerToken.B, rRes.GasFeePolicy.GasPerToken.B)

	require.Equal(t, sRes.GasFeePolicy.ValidatorFeeShare, rRes.GasFeePolicy.ValidatorFeeShare)
}
