package client

import (
	"encoding/json"
	"net/http"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/reqstatus"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

// RequestReceipt fetches the processing status of a request.
func (c *WaspClient) RequestReceipt(chainID *iscp.ChainID, reqID iscp.RequestID) (*string, error) {
	var res *string
	if err := c.do(http.MethodGet, routes.RequestReceipt(chainID.String(), reqID.String()), nil, res); err != nil {
		return nil, err
	}
	return res, nil
}

// WaitUntilRequestProcessed blocks until the request has been processed by the node
func (c *WaspClient) WaitUntilRequestProcessed(chainID *iscp.ChainID, reqID iscp.RequestID, timeout time.Duration) (*iscp.Receipt, error) {
	if timeout == 0 {
		timeout = reqstatus.WaitRequestProcessedDefaultTimeout
	}
	var res model.RequestReceiptResponse
	err := c.do(
		http.MethodGet,
		routes.WaitRequestProcessed(chainID.String(), reqID.String()),
		&model.WaitRequestProcessedParams{Timeout: timeout},
		&res,
	)
	if err != nil {
		return nil, err
	}
	var receipt iscp.Receipt
	err = json.Unmarshal([]byte(res.Receipt), &receipt)
	if err != nil {
		return nil, err
	}
	return &receipt, nil
}

// WaitUntilAllRequestsProcessed blocks until all requests in the given transaction have been processed
// by the node
func (c *WaspClient) WaitUntilAllRequestsProcessed(chainID *iscp.ChainID, tx *iotago.Transaction, timeout time.Duration) ([]*iscp.Receipt, error) {
	reqs, err := iscp.RequestsInTransaction(tx)
	if err != nil {
		return nil, err
	}
	ret := make([]*iscp.Receipt, len(reqs))
	for i, req := range reqs[*chainID] {
		receipt, err := c.WaitUntilRequestProcessed(chainID, req.ID(), timeout)
		if err != nil {
			return nil, err
		}
		ret[i] = receipt
	}
	return ret, nil
}
