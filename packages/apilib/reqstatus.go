package apilib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/plugins/webapi/stateapi"
	"net/http"
)

func QueryRequestProcessingStatusMulti(host string, addr *address.Address, reqs []sctransaction.RequestId) (map[sctransaction.RequestId]bool, error) {
	query := stateapi.ReqStateRequest{
		Address:    addr.String(),
		RequestIds: make([]string, len(reqs)),
	}
	for i := range query.RequestIds {
		query.RequestIds[i] = reqs[i].ToBase58()
	}
	data, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("http://%s/sc/state/request", host)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	var queryResponse stateapi.ReqStateResponse
	err = json.NewDecoder(resp.Body).Decode(&queryResponse)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK || queryResponse.Error != "" {
		return nil, fmt.Errorf("sc/state/request returned code %d: %s", resp.StatusCode, queryResponse.Error)
	}
	ret := make(map[sctransaction.RequestId]bool)
	for reqIdStr, v := range queryResponse.Requests {
		reqid, err := sctransaction.RequestIdFromBase58(reqIdStr)
		if err != nil {
			return nil, err
		}
		ret[*reqid] = v
	}
	return ret, nil
}

func IsRequestProcessed(host string, addr *address.Address, reqid sctransaction.RequestId) (bool, error) {
	m, err := QueryRequestProcessingStatusMulti(host, addr, []sctransaction.RequestId{reqid})
	if err != nil {
		return false, err
	}
	_, ok := m[reqid]
	return ok, nil
}
