package apilib

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/plugins/webapi/admapi"
	"github.com/iotaledger/wasp/plugins/webapi/stateapi"
)

func DumpSCState(host string, scAddress string) (*admapi.DumpSCStateResponse, error) {
	url := fmt.Sprintf("http://%s/adm/dumpscstate/%s", host, scAddress)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response status %d", resp.StatusCode)
	}
	var result admapi.DumpSCStateResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	if result.Err != "" {
		return nil, errors.New(result.Err)
	}
	return &result, nil
}

var ErrStateNotFound = errors.New("State not found")

func QuerySCState(host string, query *stateapi.QueryRequest) (map[kv.Key]*stateapi.QueryResult, error) {
	url := fmt.Sprintf("http://%s/sc/state/query", host)
	data, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	var queryResponse stateapi.QueryResponse
	err = json.NewDecoder(resp.Body).Decode(&queryResponse)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrStateNotFound
	}
	if resp.StatusCode != http.StatusOK || queryResponse.Error != "" {
		return nil, fmt.Errorf("sc/state/query returned code %d: %s", resp.StatusCode, queryResponse.Error)
	}

	results := make(map[kv.Key]*stateapi.QueryResult)
	for _, r := range queryResponse.Results {
		results[kv.Key(r.Key)] = r
	}
	return results, nil
}
