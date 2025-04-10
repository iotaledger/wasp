package webapi_validation

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/tools/stardust-migration/webapi-validation/base"
)

type CoreBlockLogValidation struct {
	base.ValidationContext
	client base.BlockLogClientWrapper
}

func NewCoreBlockLogValidation(validationContext base.ValidationContext) CoreBlockLogValidation {
	return CoreBlockLogValidation{
		ValidationContext: validationContext,
		client:            base.BlockLogClientWrapper{ValidationContext: validationContext},
	}
}

func (c *CoreBlockLogValidation) Validate(stateIndex uint32) {
	c.validateBlockInfo(stateIndex)
	c.validateRequestIDsInBlock(stateIndex)
}

func (c *CoreBlockLogValidation) validateRequestIDsInBlock(stateIndex uint32) {
	sRes, rRes := c.client.BlocklogGetRequestIDsForBlock(stateIndex)
	// Stardust supported multiple requests per output. This required us to add a "digest" at the end of each request which was formed round about like so: <OutputID;RequestIndex;>
	// In Rebased, RequestIDs are unique per request, letting us drop the digest.
	// To make the comparison simpler, drop the last two bytes and write them into a new array which then can be compared.

	stardustRequestIDs := make([]string, len(sRes.RequestIds))
	for i, requestID := range sRes.RequestIds {
		decodedRequestID := hexutil.MustDecode(requestID)
		require.Len(base.T, decodedRequestID, 34) // 34 == requestID(32)+Digest(2)
		stardustRequestIDs[i] = hexutil.Encode(decodedRequestID[:32])
	}

	require.EqualValues(base.T, stardustRequestIDs, rRes.RequestIds)
}

func (c *CoreBlockLogValidation) validateBlockInfo(stateIndex uint32) {
	sRes, rRes := c.client.BlocklogGetBlockInfo(stateIndex)

	require.Equal(base.T, sRes.BlockIndex, rRes.BlockIndex)
	require.Equal(base.T, sRes.GasBurned, rRes.GasBurned)
	require.Equal(base.T, sRes.GasFeeCharged, rRes.GasFeeCharged)
	require.Equal(base.T, sRes.NumOffLedgerRequests, rRes.NumOffLedgerRequests)
	require.Equal(base.T, sRes.NumSuccessfulRequests, rRes.NumSuccessfulRequests)
	require.Equal(base.T, sRes.TotalRequests, rRes.TotalRequests)
	require.Equal(base.T, sRes.Timestamp, rRes.Timestamp)
	// PreviousAliasOutput / PreviousAnchor omitted, as the Blocks have been rebuilt, and contain different IDs.
}
