package clients

import (
	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

func JSONDictToAPIJSONDict(jsonDict dict.JSONDict) apiclient.JSONDict {
	apiJsonDict := apiclient.NewJSONDict()

	for k, v := range jsonDict.Items {
		apiJsonDict.Items[k] = apiclient.Item{
			Key:   v.Key,
			Value: v.Value,
		}
	}

	return *apiJsonDict
}

func APIJsonDictToJSONDict(apiJsonDict apiclient.JSONDict) dict.JSONDict {
	jsonDict := dict.JSONDict{
		Items: make([]dict.Item, len(apiJsonDict.Items)),
	}

	for k, v := range apiJsonDict.Items {
		jsonDict.Items[k] = dict.Item{
			Key:   v.Key,
			Value: v.Value,
		}
	}

	return jsonDict
}

func APIJsonDictToDict(apiJsonDict apiclient.JSONDict) (dict.Dict, error) {
	jsonDict := APIJsonDictToJSONDict(apiJsonDict)

	return dict.FromJSONDict(jsonDict)
}