package client

import (
	"net/http"
	"time"

	"github.com/iotaledger/wasp/packages/coret"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

func RequestStatusRoute(chainId string, reqId string) string {
	return "chain/" + chainId + "/request/" + reqId + "/status"
}

func WaitRequestProcessedRoute(chainId string, reqId string) string {
	return "chain/" + chainId + "/request/" + reqId + "/wait"
}

type WaitRequestProcessedParams struct {
	Timeout time.Duration
}

type RequestStatusResponse struct {
	IsProcessed bool
}

func (c *WaspClient) RequestStatus(chainId *coret.ChainID, reqId *coret.RequestID) (*RequestStatusResponse, error) {
	res := &RequestStatusResponse{}
	if err := c.do(http.MethodGet, RequestStatusRoute(chainId.String(), reqId.Base58()), nil, res); err != nil {
		return nil, err
	}
	return res, nil
}

const WaitRequestProcessedDefaultTimeout = 30 * time.Second

func (c *WaspClient) WaitUntilRequestProcessed(chainId *coret.ChainID, reqId *coret.RequestID, timeout time.Duration) error {
	if timeout == 0 {
		timeout = WaitRequestProcessedDefaultTimeout
	}
	if err := c.do(
		http.MethodGet,
		WaitRequestProcessedRoute(chainId.String(), reqId.Base58()),
		&WaitRequestProcessedParams{Timeout: timeout},
		nil,
	); err != nil {
		return err
	}
	return nil
}

func (c *WaspClient) WaitUntilAllRequestsProcessed(tx *sctransaction.Transaction, timeout time.Duration) error {
	for i, req := range tx.Requests() {
		chainId := req.Target().ChainID()
		reqId := coret.NewRequestID(tx.ID(), uint16(i))
		if err := c.WaitUntilRequestProcessed(&chainId, &reqId, timeout); err != nil {
			return err
		}
	}
	return nil
}
