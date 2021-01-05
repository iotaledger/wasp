package client

import (
	"net/http"
	"time"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

func (c *WaspClient) RequestStatus(chainId *coretypes.ChainID, reqId *coretypes.RequestID) (*model.RequestStatusResponse, error) {
	res := &model.RequestStatusResponse{}
	if err := c.do(http.MethodGet, routes.RequestStatus(chainId.String(), reqId.Base58()), nil, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *WaspClient) WaitUntilRequestProcessed(chainId *coretypes.ChainID, reqId *coretypes.RequestID, timeout time.Duration) error {
	if timeout == 0 {
		timeout = model.WaitRequestProcessedDefaultTimeout
	}
	if err := c.do(
		http.MethodGet,
		routes.WaitRequestProcessed(chainId.String(), reqId.Base58()),
		&model.WaitRequestProcessedParams{Timeout: timeout},
		nil,
	); err != nil {
		return err
	}
	return nil
}

func (c *WaspClient) WaitUntilAllRequestsProcessed(tx *sctransaction.Transaction, timeout time.Duration) error {
	for i, req := range tx.Requests() {
		chainId := req.Target().ChainID()
		reqId := coretypes.NewRequestID(tx.ID(), uint16(i))
		if err := c.WaitUntilRequestProcessed(&chainId, &reqId, timeout); err != nil {
			return err
		}
	}
	return nil
}
