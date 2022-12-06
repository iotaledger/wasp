package client

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

func (c *WaspClient) CallView(chainID *isc.ChainID, hContract isc.Hname, functionName string, args dict.Dict) (dict.Dict, error) {
	arguments := args
	if arguments == nil {
		arguments = dict.Dict(nil)
	}
	var res dict.Dict
	err := c.do(http.MethodPost, routes.CallViewByName(chainID.String(), hContract.String(), functionName), arguments, &res)
	return res, err
}

func (c *WaspClient) CallViewByHname(chainID *isc.ChainID, hContract, hFunction isc.Hname, args dict.Dict) (dict.Dict, error) {
	arguments := args
	if arguments == nil {
		arguments = dict.Dict(nil)
	}
	var res dict.Dict
	err := c.do(http.MethodPost, routes.CallViewByHname(chainID.String(), hContract.String(), hFunction.String()), arguments, &res)
	return res, err
}
