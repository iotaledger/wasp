package apilib

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"net/http"
	"time"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/plugins/webapi/admapi"
	"github.com/iotaledger/wasp/plugins/webapi/stateapi"
)

func DumpSCState(host string, scAddress string) (*admapi.DumpSCStateResponse, error) {
	url := fmt.Sprintf("http://%s/adm/sc/%s/dumpstate", host, scAddress)
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

type QuerySCStateResult struct {
	Queries map[kv.Key]*stateapi.QueryResult
	// returned only when QueryGeneralData = true
	StateIndex uint32
	Timestamp  time.Time
	StateHash  *hashing.HashValue
	StateTxId  *valuetransaction.ID
	Requests   []*sctransaction.RequestId
}

func QuerySCState(host string, query *stateapi.QueryRequest) (*QuerySCStateResult, error) {
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
	// check if all query results arrived
	for _, k := range query.Query {
		if _, ok := results[kv.Key(k.Key)]; !ok {
			return nil, fmt.Errorf("inconsistency: wrong response")
		}
	}
	ret := &QuerySCStateResult{
		Queries: results,
	}
	if !query.QueryGeneralData {
		// only returning queries, no general tx data
		return ret, nil
	}
	// general data
	stateHash, err := hashing.HashValueFromBase58(queryResponse.StateHash)
	if err != nil {
		return nil, err
	}
	stateTxId, err := valuetransaction.IDFromBase58(queryResponse.StateTxId)
	if err != nil {
		return nil, err
	}
	reqIds := make([]*sctransaction.RequestId, len(queryResponse.Requests))
	for i := range reqIds {
		reqIds[i], err = sctransaction.RequestIdFromBase58(queryResponse.Requests[i])
		if err != nil {
			return nil, err
		}
	}
	ret.StateIndex = queryResponse.StateIndex
	ret.Timestamp = time.Unix(0, queryResponse.Timestamp)
	ret.StateHash = &stateHash
	ret.StateTxId = &stateTxId
	ret.Requests = reqIds
	return ret, nil
}
