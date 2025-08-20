package chainclient

import (
	"context"

	"github.com/iotaledger/wasp/v2/clients/apiextensions"
	"github.com/iotaledger/wasp/v2/packages/isc"
)

// CallView sends a request to call a view function of a given contract, and returns the result of the call
func (c *Client) CallView(ctx context.Context, msg isc.Message, blockNumberOrHash ...string) (isc.CallResults, error) {
	viewCall := apiextensions.CallViewReq(msg)
	if len(blockNumberOrHash) > 0 {
		viewCall.Block = &blockNumberOrHash[0]
	}

	return apiextensions.CallView(ctx, c.WaspClient, viewCall)
}
