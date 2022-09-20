package client

import (
	"github.com/iotaledger/wasp/packages/webapi/v1/model"
	"github.com/iotaledger/wasp/packages/webapi/v1/routes"
	"net/http"

	"github.com/iotaledger/wasp/packages/isc"
)

func (c *WaspClient) PostOffLedgerRequest(chainID isc.ChainID, req isc.OffLedgerRequest) error {
	data := model.OffLedgerRequestBody{
		Request: model.NewBytes(req.Bytes()),
	}
	return c.do(http.MethodPost, routes.NewRequest(chainID.String()), data, nil)
}
