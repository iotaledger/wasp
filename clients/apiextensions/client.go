package apiextensions

import (
	"context"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/packages/kv/dict"
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

func CallView(context context.Context, client *apiclient.APIClient, chainID string, request apiclient.ContractCallViewRequest) (dict.Dict, error) {
	result, _, err := client.ChainsApi.
		CallView(context, chainID).
		ContractCallViewRequest(request).
		Execute()
	if err != nil {
		return nil, err
	}

	dictResult, err := APIJsonDictToDict(*result)

	return dictResult, err
}
