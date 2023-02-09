package apiextensions

import (
	"context"
	"net/url"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

func WaspAPIClientByHostName(hostname string) (*apiclient.APIClient, error) {
	parsed, err := url.Parse(hostname)
	if err != nil {
		return nil, err
	}

	config := apiclient.NewConfiguration()
	config.Host = parsed.Host
	config.Scheme = parsed.Scheme

	return apiclient.NewAPIClient(config), nil
}

func CallView(context context.Context, client *apiclient.APIClient, request apiclient.ContractCallViewRequest) (dict.Dict, error) {
	result, _, err := client.RequestsApi.
		CallView(context).
		ContractCallViewRequest(request).
		Execute()
	if err != nil {
		return nil, err
	}

	dictResult, err := APIJsonDictToDict(*result)

	return dictResult, err
}
