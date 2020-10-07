package client

import (
	"net/http"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/client/statequery"
)

func StateQueryRoute(scAddr string) string {
	return "sc/" + scAddr + "/state/query"
}

func (c *WaspClient) StateQuery(scAddress *address.Address, query *statequery.Request) (*statequery.Results, error) {
	res := &statequery.Results{}
	if err := c.do(http.MethodGet, StateQueryRoute(scAddress.String()), query, res); err != nil {
		return nil, err
	}
	return res, nil
}
