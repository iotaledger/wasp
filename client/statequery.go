// +build ignore

package client

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/webapi/model/statequery"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

// StateQuery queries the chain state, and returns the result of the query.
func (c *WaspClient) StateQuery(chainID *coretypes.ChainID, query *statequery.Request) (*statequery.Results, error) {
	res := &statequery.Results{}
	if err := c.do(http.MethodGet, routes.StateQuery(chainID.Base58()), query, res); err != nil {
		return nil, err
	}
	return res, nil
}
