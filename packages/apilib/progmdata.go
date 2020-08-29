package apilib

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/progmeta"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/util/multicall"
	"github.com/iotaledger/wasp/plugins/webapi/admapi"
	"github.com/iotaledger/wasp/plugins/webapi/misc"
	"net/http"
	"time"
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

// GetProgramMetadata calls node to get ProgramMetadata by program hash
func GetProgramMetadata(host string, progHash *hashing.HashValue) (*progmeta.ProgramMetadata, error) {
	data, err := json.Marshal(&admapi.GetProgramMetadataRequest{
		ProgramHash: progHash.String(),
	})
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("http://%s/adm/getprogrammetadata", host)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response status %d", resp.StatusCode)
	}
	var dresp admapi.GetProgramMetadataResponse
	err = json.NewDecoder(resp.Body).Decode(&dresp)
	if err != nil {
		return nil, err
	}
	if dresp.Error != "" {
		return nil, errors.New(dresp.Error)
	}
	if !dresp.ExistsMetadata {
		return nil, nil
	}
	ph, err := hashing.HashValueFromBase58(dresp.ProgramHash)
	if err != nil {
		return nil, err
	}
	return &progmeta.ProgramMetadata{
		ProgramHash:   ph,
		Location:      dresp.Location,
		VMType:        dresp.VMType,
		Description:   dresp.Description,
		CodeAvailable: dresp.ExistsCode,
	}, nil
}

// CheckProgramMetadata checks if metadata exists in hosts and is consistent
// return program meta data from the first host if all consistent, otherwise nil
func CheckProgramMetadata(hosts []string, progHash *hashing.HashValue) (*progmeta.ProgramMetadata, error) {
	funs := make([]func() error, len(hosts))
	mdata := make([]*progmeta.ProgramMetadata, len(hosts))
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
func consistentProgramMetadata(md1, md2 *progmeta.ProgramMetadata) bool {
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
	if md1.Location != md2.Location {
		return false
	}
	return true
}
