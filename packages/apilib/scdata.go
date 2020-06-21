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

// PutSCData calls node to write BootupData record
func PutSCData(host string, bd registry.BootupData) error {
	data, err := json.Marshal(&admapi.BootupDataJsonable{
		Address:        bd.Address.String(),
		OwnerAddress:   bd.OwnerAddress.String(),
		Color:          bd.Color.String(),
		CommitteeNodes: bd.CommitteeNodes,
		AccessNodes:    bd.AccessNodes,
	})
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

// GetSCData calls node to get BootupData record by address
func GetSCData(host string, addr *address.Address) (*registry.BootupData, bool, error) {
	data, err := json.Marshal(&admapi.GetBootupDataRequest{
		Address: addr.String(),
	})
	if err != nil {
		return nil, false, err
	}
	url := fmt.Sprintf("http://%s/adm/getscdata", host)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, false, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("response status %d", resp.StatusCode)
	}
	var dresp admapi.GetBootupDataResponse
	err = json.NewDecoder(resp.Body).Decode(&dresp)
	if err != nil {
		return nil, false, err
	}
	if dresp.Error != "" {
		return nil, false, errors.New(dresp.Error)
	}
	if !dresp.Exists {
		return nil, false, nil
	}
	ret := &registry.BootupData{
		CommitteeNodes: dresp.CommitteeNodes,
	}
	if ret.Address, err = address.FromBase58(dresp.Address); err != nil {
		return nil, false, err
	}

	return ret, true, nil
}

// gets list of all SCs from the node
func GetSCList(url string) ([]address.Address, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s/adm/getsclist", url))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response status %d", resp.StatusCode)
	}
	var lresp admapi.GetScAddressesResponse
	err = json.NewDecoder(resp.Body).Decode(&lresp)
	if err != nil {
		return nil, err
	}
	if lresp.Error != "" {
		return nil, errors.New(lresp.Error)
	}
	ret := make([]address.Address, len(lresp.Addresses))
	for i, addrstr := range lresp.Addresses {
		addr, err := address.FromBase58(addrstr)
		if err == nil {
			ret[i] = addr
		}
	}
	return ret, nil
}
