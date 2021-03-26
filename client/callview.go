package client

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

// CallView sends a request to call a view function of a given contract, and returns the result of the call
func (c *WaspClient) CallView(chainID coretypes.ChainID, hContract coretypes.Hname, functionName string, args ... dict.Dict) (dict.Dict, error) {
	arguments := dict.Dict(nil)
	if args != nil && len(args) != 0 {
		arguments = args[0]
	}
	var res dict.Dict
	contractID := coretypes.NewAgentID(chainID.AsAddress(), hContract)
	if err := c.do(http.MethodGet, routes.CallView(contractID.Base58(), functionName), arguments, &res); err != nil {
		return nil, err
	}
	return res, nil
}
