package chainclient

import (
	"context"

	"github.com/iotaledger/wasp/clients"
	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// CallView sends a request to call a view function of a given contract, and returns the result of the call
func (c *Client) CallView(ctx context.Context, hContract isc.Hname, functionName string, args dict.Dict) (dict.Dict, error) {

	request := apiclient.ContractCallViewRequest{
		ChainId:       c.ChainID.String(),
		ContractHName: hContract.String(),
		FunctionName:  functionName,
		Arguments:     *clients.JSONDictToAPIJSONDict(args.JSONDict()),
	}

	result, _, err := c.WaspClient.RequestsApi.
		CallView(ctx).
		ContractCallViewRequest(request).
		Execute()

	dictResult, err := dict.FromJSONDict(clients.APIJsonDictToJSONDict(*result))

	return dictResult, err
}
