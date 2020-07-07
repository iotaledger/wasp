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

func QuerySCState(host string, scAddress string, keys []kv.Key) (kv.Map, error) {
	url := fmt.Sprintf("http://%s/sc/state/query", host)
	data, err := json.Marshal(&stateapi.QueryStateRequest{
		Address: scAddress,
		Keys:    toBytes(keys),
	})
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	var result stateapi.QueryStateResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK || result.Error != "" {
		return nil, fmt.Errorf("sc/state/query returned code %d: %s", resp.StatusCode, result.Error)
	}
	return toMap(result.Values), nil
}

func toBytes(keys []kv.Key) [][]byte {
	ret := make([][]byte, 0)
	for _, v := range keys {
		ret = append(ret, []byte(v))
	}
	return ret
}

func toMap(values []stateapi.KeyValuePair) kv.Map {
	m := kv.NewMap()
	for _, entry := range values {
		if len(entry.Value) > 0 {
			m.Set(kv.Key(entry.Key), entry.Value)
		}
	}
	return m
}
