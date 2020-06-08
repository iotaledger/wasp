package apilib

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/iotaledger/wasp/plugins/webapi/admapi"
	"github.com/iotaledger/wasp/plugins/webapi/misc"
	"net/http"
)

func ActivateSC(host, addr string) error {
	data, err := json.Marshal(&admapi.ActivateSCRequest{
		Address: addr,
	})
	if err != nil {
		return err
	}
	url := fmt.Sprintf("http://%s/adm/activatesc", host)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("response status %d", resp.StatusCode)
	}
	var aresp misc.SimpleResponse
	err = json.NewDecoder(resp.Body).Decode(&aresp)
	if err != nil {
		return err
	}
	if aresp.Error != "" {
		return errors.New(aresp.Error)
	}
	return nil
}
