package client

import (
	"net/http"
	"time"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

func RequestStatusRoute(chainId string, reqId string) string {
	return "chain/" + chainId + "/request/" + reqId + "/status"
}

func WaitRequestProcessedRoute(chainId string, reqId string) string {
	return "chain/" + chainId + "/request/" + reqId + "/wait"
}

type WaitRequestProcessedParams struct {
	Timeout time.Duration `swagger:"desc(Timeout in nanoseconds),default(30 seconds)"`
}

type RequestStatusResponse struct {
	IsProcessed bool `swagger:"desc(True if the request has been processed)"`
}

func (c *WaspClient) RequestStatus(chainId *coretypes.ChainID, reqId *coretypes.RequestID) (*RequestStatusResponse, error) {
	res := &RequestStatusResponse{}
	if err := c.do(http.MethodGet, RequestStatusRoute(chainId.String(), reqId.Base58()), nil, res); err != nil {
		return nil, err
	}
	return res, nil
}

const WaitRequestProcessedDefaultTimeout = 30 * time.Second

func (c *WaspClient) WaitUntilRequestProcessed(chainId *coretypes.ChainID, reqId *coretypes.RequestID, timeout time.Duration) error {
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
		reqId := coretypes.NewRequestID(tx.ID(), uint16(i))
		if err := c.WaitUntilRequestProcessed(&chainId, &reqId, timeout); err != nil {
			return err
		}
	}
	return nil
}
