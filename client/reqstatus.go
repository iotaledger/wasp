package client

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"net/http"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
)

func RequestStatusRoute(scAddr string, reqId string) string {
	return "sc/" + scAddr + "/request/" + reqId + "/status"
}

type RequestStatusResponse struct {
	IsProcessed bool
}

func (c *WaspClient) RequestStatus(scAddr *address.Address, requestId *coretypes.RequestID) (*RequestStatusResponse, error) {
	res := &RequestStatusResponse{}
	if err := c.do(http.MethodGet, RequestStatusRoute(scAddr.String(), requestId.Base58()), nil, res); err != nil {
		return nil, err
	}
	return res, nil
}
