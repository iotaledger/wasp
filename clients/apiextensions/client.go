package apiextensions

import (
	"context"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/packages/isc"
)

func WaspAPIClientByHostName(hostname string) (*apiclient.APIClient, error) {
	_, err := ValidateAbsoluteURL(hostname)
	if err != nil {
		return nil, err
	}

	config := apiclient.NewConfiguration()
	config.Servers[0].URL = hostname

	return apiclient.NewAPIClient(config), nil
}

func CallView(context context.Context, client *apiclient.APIClient, chainID string, request apiclient.ContractCallViewRequest) (isc.CallArguments, error) {
	_, _, err := client.ChainsApi.
		CallView(context, chainID).
		ContractCallViewRequest(request).
		Execute()
	if err != nil {
		return nil, err
	}

	panic("we need to recompile the webapi docs to get a proper type from  result. For this ISC needs to be compilable. Right now faking a hex string")

	result := "0x0"
	resultHex, err := hexutil.Decode(result)
	if err != nil {
		return nil, err
	}

	resultArgs, err := isc.CallArgumentsFromBytes(resultHex)
	if err != nil {
		return nil, err
	}

	return resultArgs, nil
}
