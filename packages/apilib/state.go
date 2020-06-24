package apilib

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/iotaledger/wasp/plugins/webapi/admapi"
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
