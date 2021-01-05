package client

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

func (c *WaspClient) DumpSCState(scid *coretypes.ContractID) (*model.SCStateDump, error) {
	res := &model.SCStateDump{}
	if err := c.do(http.MethodGet, routes.DumpState(scid.Base58()), nil, res); err != nil {
		return nil, err
	}
	return res, nil
}
