package apilib

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/webapi/admapi"
	"github.com/iotaledger/wasp/plugins/webapi/misc"
	"net/http"
)

// PutProgramMetadata calls node to write ProgramMetadata record
func PutProgramMetadata(host string, md registry.ProgramMetadata) error {
	data, err := json.Marshal(&admapi.ProgramMetadataJsonable{
		ProgramHash: md.ProgramHash.String(),
		Location:    md.Location,
		Description: md.Description,
	})
	if err != nil {
		return err
	}
	url := fmt.Sprintf("http://%s/adm/putprogrammetadata", host)
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

// GetProgramMetadata calls node to get ProgramMetadata record by program hash
// return if exist metadata and if exist program code in the registry
func GetProgramMetadata(host string, progHash *hashing.HashValue) (*registry.ProgramMetadata, bool, bool, error) {
	data, err := json.Marshal(&admapi.GetProgramMetadataRequest{
		ProgramHash: progHash.String(),
	})
	if err != nil {
		return nil, false, false, err
	}
	url := fmt.Sprintf("http://%s/adm/getprogrammetadata", host)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, false, false, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, false, fmt.Errorf("response status %d", resp.StatusCode)
	}
	var dresp admapi.GetProgramMetadataResponse
	err = json.NewDecoder(resp.Body).Decode(&dresp)
	if err != nil {
		return nil, false, false, err
	}
	if dresp.Error != "" {
		return nil, false, false, errors.New(dresp.Error)
	}
	if !dresp.ExistsMetadata {
		return nil, false, false, nil
	}
	ph, err := hashing.HashValueFromBase58(dresp.ProgramHash)
	if err != nil {
		return nil, false, false, err
	}
	ret := &registry.ProgramMetadata{
		ProgramHash: ph,
		Location:    dresp.Location,
		Description: dresp.Description,
	}

	return ret, true, dresp.ExistsCode, nil
}
