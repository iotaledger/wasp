package apilib

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/plugins/webapi/admapi"
	"github.com/iotaledger/wasp/plugins/webapi/misc"
	"net/http"
)

// calls node  to wright SCMetaData record
func PutSCData(host string, params admapi.SCMetaDataJsonable) error {
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("http://%s/adm/putscdata", host)
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

// calls the nodes to get SCMetaData record by address
func GetSCdata(host string, scaddr *address.Address) (*admapi.SCMetaDataJsonable, error) {
	req := admapi.GetSCDataRequest{Address: scaddr}
	data, err := json.Marshal(&req)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("http://%s/adm/getscdata", host)
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
	return &dresp.SCMetaDataJsonable, err
}

// gets list of all SCs from the node
func GetSCList(url string) ([]*admapi.SCMetaDataJsonable, error) {
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
