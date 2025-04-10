package webapi_validation

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/require"
)

type CoreBlockLogValidation struct {
	ValidationContext
}

func NewCoreBlockLogValidation(validationContext ValidationContext) CoreBlockLogValidation {
	return CoreBlockLogValidation{
		ValidationContext: validationContext,
	}
}

func (c *CoreBlockLogValidation) Validate(stateIndex uint32) {
	c.validateBlockInfo(stateIndex)
	c.validateRequestIDsInBlock(stateIndex)
}

func (c *CoreBlockLogValidation) validateRequestIDsInBlock(stateIndex uint32) {
	sRes, _, err := c.sClient.CorecontractsApi.BlocklogGetRequestIDsForBlock(c.ctx, MainnetChainID, stateIndex).Block(Uint32ToString(stateIndex)).Execute()
	require.NoError(t, err)

	rRes, _, err := c.rClient.CorecontractsAPI.BlocklogGetRequestIDsForBlock(c.ctx, stateIndex).Block(Uint32ToString(stateIndex)).Execute()
	require.NoError(t, err)

	// Stardust supported multiple requests per output. This required us to add a "digest" at the end of each request which was formed round about like so: <OutputID;RequestIndex;>
	// In Rebased, RequestIDs are unique per request, letting us drop the digest.
	// To make the comparison simpler, drop the last two bytes and write them into a new array which then can be compared.

	stardustRequestIDs := make([]string, len(sRes.RequestIds))
	for i, requestID := range sRes.RequestIds {
		decodedRequestID := hexutil.MustDecode(requestID)
		require.Len(t, decodedRequestID, 34) // 34 == requestID(32)+Digest(2)
		stardustRequestIDs[i] = hexutil.Encode(decodedRequestID[:32])
	}

	require.EqualValues(t, stardustRequestIDs, rRes.RequestIds)
}

func (c *CoreBlockLogValidation) validateBlockInfo(stateIndex uint32) {
	// As the blocks are pruned except for the last 10000, it's required to tell the WebAPI to not use the latest state, but actually the state of `stateIndex`
	// This is why there is a .Block(stateIndex) there.
	rRes, _, err := c.rClient.CorecontractsAPI.BlocklogGetBlockInfo(c.ctx, stateIndex).Block(Uint32ToString(stateIndex)).Execute()
	require.NoError(t, err)

	sRes, _, err := c.sClient.CorecontractsApi.BlocklogGetBlockInfo(c.ctx, MainnetChainID, stateIndex).Block(Uint32ToString(stateIndex)).Execute()
	require.NoError(t, err)

	require.Equal(t, sRes.BlockIndex, rRes.BlockIndex)
	require.Equal(t, sRes.GasBurned, rRes.GasBurned)
	require.Equal(t, sRes.GasFeeCharged, rRes.GasFeeCharged)
	require.Equal(t, sRes.NumOffLedgerRequests, rRes.NumOffLedgerRequests)
	require.Equal(t, sRes.NumSuccessfulRequests, rRes.NumSuccessfulRequests)
	require.Equal(t, sRes.TotalRequests, rRes.TotalRequests)
	require.Equal(t, sRes.Timestamp, rRes.Timestamp)
	// PreviousAliasOutput / PreviousAnchor omitted, as the Blocks have been rebuilt, and contain different IDs.
}
