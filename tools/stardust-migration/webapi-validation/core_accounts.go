package webapi_validation

import (
	"github.com/stretchr/testify/require"
)

type CoreAccountsValidation struct {
	ValidationContext
}

func NewCoreAccountsValidation(validationContext ValidationContext) CoreAccountsValidation {
	return CoreAccountsValidation{
		ValidationContext: validationContext,
	}
}

func (c *CoreAccountsValidation) Validate(stateIndex uint32) {
	c.validateBlockInfo(stateIndex)
}

func (c *CoreAccountsValidation) validateBlockInfo(stateIndex uint32) {
	sRes, _, err := c.sClient.CorecontractsApi.BlocklogGetBlockInfo(c.ctx, MainnetChainID, stateIndex).Execute()
	require.NoError(t, err)

	rRes, _, err := c.rClient.CorecontractsAPI.BlocklogGetBlockInfo(c.ctx, stateIndex).Execute()
	require.NoError(t, err)

	require.Equal(t, sRes.BlockIndex, rRes.BlockIndex)
	require.Equal(t, sRes.GasBurned, rRes.GasBurned)
	require.Equal(t, sRes.GasFeeCharged, rRes.GasFeeCharged)
	require.Equal(t, sRes.NumOffLedgerRequests, rRes.NumOffLedgerRequests)
	require.Equal(t, sRes.NumSuccessfulRequests, rRes.NumSuccessfulRequests)
	require.Equal(t, sRes.TotalRequests, rRes.TotalRequests)
	require.Equal(t, sRes.Timestamp, rRes.Timestamp)
}
