package client

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

// CallView sends a request to call a view function of a given contract, and returns the result of the call
func (c *WaspClient) CallView(contractID coretypes.ContractID, fname string, arguments dict.Dict) (dict.Dict, error) {
	var res dict.Dict
	if err := c.do(http.MethodGet, routes.CallView(contractID.Base58(), fname), arguments, &res); err != nil {
		return nil, err
	}
	return res, nil
}
