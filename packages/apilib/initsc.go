package apilib

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/plugins/webapi/admapi"
	"github.com/iotaledger/wasp/plugins/webapi/misc"
	"github.com/mr-tron/base58"
	"net/http"
)

// sending origin batch to one of committee nodes
func PostOriginBatch(node string, addr *address.Address, batch state.Batch) error {
	batchBytes, err := util.Bytes(batch)
	if err != nil {
		return err
	}
	data, err := json.Marshal(&admapi.InitScRequest{
		Address:   addr.String(),
		BatchData: base58.Encode(batchBytes),
	})
	if err != nil {
		return err
	}
	url := fmt.Sprintf("http://%s/adm/initsc", node)
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
