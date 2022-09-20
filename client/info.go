package client

import (
	"github.com/iotaledger/wasp/packages/webapi/v1/model"
	"github.com/iotaledger/wasp/packages/webapi/v1/routes"
	"net/http"
)

// Info fetches general information about the node.
func (c *WaspClient) Info() (*model.InfoResponse, error) {
	res := &model.InfoResponse{}
	if err := c.do(http.MethodGet, routes.Info(), nil, res); err != nil {
		return nil, err
	}
	return res, nil
}
