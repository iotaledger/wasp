package apilib

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/util/multicall"
	"github.com/iotaledger/wasp/plugins/webapi/admapi"
)

// PutProgramMetadata calls node to write program code and ProgramMetadata record
func PutProgram(host string, vmType string, description string, code []byte) (*hashing.HashValue, error) {
	data, err := json.Marshal(&admapi.PutProgramRequest{
		ProgramMetadata: admapi.ProgramMetadata{
			VMType:      vmType,
			Description: description,
		},
		Code: code,
	})
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("http://%s/adm/program", host)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	var result admapi.PutProgramResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	if result.Error != "" {
		return nil, errors.New(result.Error)
	}
	hash, err := hashing.HashValueFromBase58(result.ProgramHash)
	if err != nil {
		return nil, err
	}
	return &hash, nil
}

// GetProgramMetadata calls node to get ProgramMetadata by program hash
func GetProgramMetadata(host string, progHash *hashing.HashValue) (*registry.ProgramMetadata, error) {
	url := fmt.Sprintf("http://%s/adm/program/%s", host, progHash.String())
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	var dresp admapi.GetProgramMetadataResponse
	err = json.NewDecoder(resp.Body).Decode(&dresp)
	if err != nil {
		return nil, err
	}
	if dresp.Error != "" {
		return nil, errors.New(dresp.Error)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response status %d", resp.StatusCode)
	}
	return &registry.ProgramMetadata{
		ProgramHash: *progHash,
		VMType:      dresp.VMType,
		Description: dresp.Description,
	}, nil
}

// CheckProgramMetadata checks if metadata exists in hosts and is consistent
// return program meta data from the first host if all consistent, otherwise nil
func CheckProgramMetadata(hosts []string, progHash *hashing.HashValue) (*registry.ProgramMetadata, error) {
	funs := make([]func() error, len(hosts))
	mdata := make([]*registry.ProgramMetadata, len(hosts))
	for i, host := range hosts {
		var err error
		h := host
		idx := i
		funs[i] = func() error {
			mdata[idx], err = GetProgramMetadata(h, progHash)
			return err
		}
	}
	succ, errs := multicall.MultiCall(funs, 1*time.Second)
	if !succ {
		return nil, multicall.WrapErrors(errs)
	}
	errInconsistent := fmt.Errorf("non existent or inconsistent program metadata for program hash %s", progHash.String())
	for _, md := range mdata {
		if !consistentProgramMetadata(mdata[0], md) {
			return nil, errInconsistent
		}
	}
	return mdata[0], nil
}

// consistentProgramMetadata does not check if code exists
func consistentProgramMetadata(md1, md2 *registry.ProgramMetadata) bool {
	if md1 == nil || md2 == nil {
		return false
	}
	if md1.ProgramHash != md2.ProgramHash {
		return false
	}
	if md1.VMType != md2.VMType {
		return false
	}
	if md1.Description != md2.Description {
		return false
	}
	return true
}
