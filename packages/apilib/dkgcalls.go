package apilib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/plugins/webapi/dkgapi"
	"github.com/pkg/errors"
)

func callNewKey(netLoc string, params dkgapi.NewDKSRequest) (*dkgapi.NewDKSResponse, error) {
	data, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("http://%s/adm/newdks", netLoc)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	result := &dkgapi.NewDKSResponse{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	if result.Err == "" {
		return result, nil
	}
	return nil, errors.New(result.Err)
}

func callAggregate(netLoc string, params dkgapi.AggregateDKSRequest) (*dkgapi.AggregateDKSResponse, error) {
	data, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("http://%s/adm/aggregatedks", netLoc)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	result := &dkgapi.AggregateDKSResponse{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	if result.Err == "" {
		return result, nil
	}
	return nil, errors.New(result.Err)
}

func callCommit(netloc string, params dkgapi.CommitDKSRequest) (*address.Address, error) {
	data, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("http://%s/adm/commitdks", netloc)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	result := &dkgapi.CommitDKSResponse{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	if result.Err == "" {
		addrRet, err := address.FromBase58(result.Address)
		if err != nil {
			return nil, err
		}
		return &addrRet, nil
	}
	return nil, errors.New(result.Err)
}

func callGetPubKeyInfo(netLoc string, params dkgapi.GetPubKeyInfoRequest) *dkgapi.GetPubKeyInfoResponse {
	data, err := json.Marshal(params)
	if err != nil {
		return &dkgapi.GetPubKeyInfoResponse{Err: err.Error()}
	}
	url := fmt.Sprintf("http://%s/adm/getpubkeyinfo", netLoc)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return &dkgapi.GetPubKeyInfoResponse{Err: err.Error()}
	}
	result := &dkgapi.GetPubKeyInfoResponse{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return &dkgapi.GetPubKeyInfoResponse{Err: err.Error()}
	}
	return result
}

func callExportDKShare(netLoc string, params dkgapi.ExportDKShareRequest) (string, error) {
	data, err := json.Marshal(params)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("http://%s/adm/exportdkshare", netLoc)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	result := &dkgapi.ExportDKShareResponse{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("exportdkshare returned code %d: %s", resp.StatusCode, result.Err)
	}
	return result.DKShare, err
}

func callImportDKShare(netLoc string, params dkgapi.ImportDKShareRequest) error {
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("http://%s/adm/importdkshare", netLoc)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	result := &dkgapi.ImportDKShareResponse{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("importdkshare returned code %d: %s", resp.StatusCode, result.Err)
	}
	return err
}
