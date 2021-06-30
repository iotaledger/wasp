package client

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/coretypes/chainid"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

// CallView sends a request to call a view function of a given contract, and returns the result of the call
func (c *WaspClient) CallView(chainID chainid.ChainID, hContract coretypes.Hname, functionName string, args ...dict.Dict) (dict.Dict, error) {
	arguments := dict.Dict(nil)
	if len(args) != 0 {
		arguments = args[0]
	}
	var res dict.Dict
	if err := c.do(http.MethodGet, routes.CallView(chainID.Base58(), hContract.String(), functionName), arguments, &res); err != nil {
		return nil, err
	}
	return res, nil
}
