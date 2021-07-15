package client

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/request"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

func (c *WaspClient) PostOffLedgerRequest(chainID *coretypes.ChainID, req *request.RequestOffLedger) error {
	data := model.OffLedgerRequestBody{
		Request: model.NewBytes(req.Bytes()),
	}
	return c.do("POST", routes.NewRequest(chainID.Base58()), data, nil)
}
