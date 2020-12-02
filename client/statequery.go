package client

import (
	"net/http"

	"github.com/iotaledger/wasp/client/statequery"
	"github.com/iotaledger/wasp/packages/coretypes"
)

func StateQueryRoute(chainID string) string {
	return "chain/" + chainID + "/state/query"
}

func (c *WaspClient) StateQuery(chainID *coretypes.ChainID, query *statequery.Request) (*statequery.Results, error) {
	res := &statequery.Results{}
	if err := c.do(http.MethodGet, StateQueryRoute(chainID.String()), query, res); err != nil {
		return nil, err
	}
	return res, nil
}
