package client

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/request"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

func (c *WaspClient) PostOffLedgerRequest(chainID *iscp.ChainID, req *request.OffLedger) error {
	data := model.OffLedgerRequestBody{
		Request: model.NewBytes(req.Bytes()),
	}
	return c.do("POST", routes.NewRequest(chainID.Base58()), data, nil)
}
