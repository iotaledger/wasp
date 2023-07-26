package chainclient

import (
	"context"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// CallView sends a request to call a view function of a given contract, and returns the result of the call
func (c *Client) CallView(ctx context.Context, hContract isc.Hname, functionName string, args dict.Dict, blockNumberOrHash ...string) (dict.Dict, error) {
	viewCall := apiclient.ContractCallViewRequest{
		ContractHName: hContract.String(),
		FunctionName:  functionName,
		Arguments:     apiextensions.JSONDictToAPIJSONDict(args.JSONDict()),
	}
	if len(blockNumberOrHash) > 0 {
		viewCall.Block = &blockNumberOrHash[0]
	}

	return apiextensions.CallView(ctx, c.WaspClient, c.ChainID.String(), viewCall)
}
