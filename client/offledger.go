package client

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

func (c *WaspClient) PostOffLedgerRequest(chainID *isc.ChainID, req isc.OffLedgerRequest) error {
	data := model.OffLedgerRequestBody{
		Request: model.NewBytes(req.Bytes()),
	}
	return c.do(http.MethodPost, routes.NewRequest(chainID.String()), data, nil)
}
