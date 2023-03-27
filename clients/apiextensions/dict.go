package apiextensions

import (
	"context"
	"fmt"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

func JSONDictToAPIJSONDict(jsonDict dict.JSONDict) apiclient.JSONDict {
	apiJSONDict := apiclient.NewJSONDict()
	apiJSONDict.Items = make([]apiclient.Item, len(jsonDict.Items))

	for k, v := range jsonDict.Items {
		apiJSONDict.Items[k] = apiclient.Item{
			Key:   v.Key,
			Value: v.Value,
		}
	}

	return *apiJSONDict
}

func DictToAPIJsonDict(dict dict.Dict) apiclient.JSONDict {
	return JSONDictToAPIJSONDict(dict.JSONDict())
}

func APIJsonDictToJSONDict(apiJSONDict apiclient.JSONDict) dict.JSONDict {
	jsonDict := dict.JSONDict{
		Items: make([]dict.Item, len(apiJSONDict.Items)),
	}

	for k, v := range apiJSONDict.Items {
		jsonDict.Items[k] = dict.Item{
			Key:   v.Key,
			Value: v.Value,
		}
	}

	return jsonDict
}

func APIJsonDictToDict(apiJSONDict apiclient.JSONDict) (dict.Dict, error) {
	jsonDict := APIJsonDictToJSONDict(apiJSONDict)

	return dict.FromJSONDict(jsonDict)
}

func APIWaitUntilAllRequestsProcessed(client *apiclient.APIClient, chainID isc.ChainID, tx *iotago.Transaction, timeout time.Duration) ([]*apiclient.ReceiptResponse, error) {
	reqs, err := isc.RequestsInTransaction(tx)
	if err != nil {
		return nil, err
	}
	ret := make([]*apiclient.ReceiptResponse, len(reqs))
	for i, req := range reqs[chainID] {
		receipt, _, err := client.RequestsApi.
			WaitForRequest(context.Background(), chainID.String(), req.ID().String()).
			TimeoutSeconds(int32(timeout.Seconds())).
			Execute()
		if err != nil {
			return nil, fmt.Errorf("Error in WaitForRequest, reqID=%v: %w", req.ID(), err)
		}

		ret[i] = receipt
	}
	return ret, nil
}
