package apilib

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/webapi/admapi"
	"github.com/iotaledger/wasp/plugins/webapi/misc"
	"net/http"
)

// calls node  to wright SCData record
func PutSCData(addr string, port int, adata *registry.SCData) error {
	data, err := json.Marshal(adata)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("http://%s:%d/adm/putscdata", addr, port)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	var result misc.SimpleResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return err
	}
	if result.Error != "" {
		err = errors.New(result.Error)
	}
	return err
}

// calls the nodes to get SCData record by scid
func GetSCdata(addr string, port int, scaddr *address.Address) (*registry.SCData, error) {
	req := admapi.GetSCDataRequest{Address: scaddr}
	data, err := json.Marshal(&req)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("http://%s:%d/adm/getscdata", addr, port)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response status %d", resp.StatusCode)
	}
	var dresp admapi.GetSCDataResponse
	err = json.NewDecoder(resp.Body).Decode(&dresp)
	if err != nil {
		return nil, err
	}
	if dresp.Error != "" {
		return nil, errors.New(dresp.Error)
	}
	return &dresp.SCData, err
}

// gets list of all SCs from the node
func GetSCList(url string) ([]*registry.SCData, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s/adm/getsclist", url))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response status %d", resp.StatusCode)
	}
	var lresp admapi.GetScListResponse
	err = json.NewDecoder(resp.Body).Decode(&lresp)
	if err != nil {
		return nil, err
	}
	if lresp.Error != "" {
		return nil, errors.New(lresp.Error)
	}
	return lresp.SCDataList, nil
}
