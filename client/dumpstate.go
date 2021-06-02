package client

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

func (c *WaspClient) DumpSCState(chainID *coretypes.ChainID, hname coretypes.Hname) (*model.SCStateDump, error) {
	res := &model.SCStateDump{}
	if err := c.do(http.MethodGet, routes.DumpState(chainID.Base58(), hname.String()), nil, res); err != nil {
		return nil, err
	}
	return res, nil
}
