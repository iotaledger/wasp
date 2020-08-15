package apilib

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/util"
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

// GetProgramMetadata calls node to get ProgramMetadata record by program hash
// return if exist program code in the registry
func GetProgramMetadata(host string, progHash *hashing.HashValue) (*registry.ProgramMetadata, bool, error) {
	data, err := json.Marshal(&admapi.GetProgramMetadataRequest{
		ProgramHash: progHash.String(),
	})
	if err != nil {
		return nil, false, err
	}
	url := fmt.Sprintf("http://%s/adm/getprogrammetadata", host)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, false, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("response status %d", resp.StatusCode)
	}
	var dresp admapi.GetProgramMetadataResponse
	err = json.NewDecoder(resp.Body).Decode(&dresp)
	if err != nil {
		return nil, false, err
	}
	if dresp.Error != "" {
		return nil, false, errors.New(dresp.Error)
	}
	if !dresp.ExistsMetadata {
		return nil, false, nil
	}
	ph, err := hashing.HashValueFromBase58(dresp.ProgramHash)
	if err != nil {
		return nil, false, err
	}
	ret := &registry.ProgramMetadata{
		ProgramHash: ph,
		Location:    dresp.Location,
		Description: dresp.Description,
	}

	return ret, dresp.ExistsCode, nil
}

// CheckProgramMetadata checks if metadata exists in hosts and is consistent
func CheckProgramMetadata(hosts []string, progHash *hashing.HashValue) error {
	funs := make([]func() error, len(hosts))
	mdata := make([]*registry.ProgramMetadata, len(hosts))
	for i, host := range hosts {
		var err error
		h := host
		funs[i] = func() error {
			mdata[i], _, err = GetProgramMetadata(h, progHash)
			return err
		}
	}
	succ, errs := multicall.MultiCall(funs, 1*time.Second)
	if !succ {
		return multicall.WrapErrors(errs)
	}
	errInconsistent := fmt.Errorf("non existent or inconsistent program metadata for program hash %s", progHash.String())
	if mdata[0] == nil {
		return errInconsistent
	}
	md0, _ := util.Bytes(mdata[0])
	for _, md := range mdata {
		if md == nil {
			return errInconsistent
		}
		mdi, _ := util.Bytes(md)
		if !bytes.Equal(mdi, md0) {
			return errInconsistent
		}
	}
	return nil
}
