package client

import (
	"encoding/json"
	"net/http"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/v1/model"
	"github.com/iotaledger/wasp/packages/webapi/v1/routes"
)

// RequestReceipt fetches the processing status of a request.
func (c *WaspClient) RequestReceipt(chainID isc.ChainID, reqID isc.RequestID) (*isc.Receipt, error) {
	var res model.RequestReceiptResponse
	if err := c.do(http.MethodGet, routes.RequestReceipt(chainID.String(), reqID.String()), nil, &res); err != nil {
		return nil, err
	}
	if res.Receipt == "" {
		return nil, nil
	}
	var receipt isc.Receipt
	err := json.Unmarshal([]byte(res.Receipt), &receipt)
	if err != nil {
		return nil, err
	}
	return &receipt, nil
}

const waitRequestProcessedDefaultTimeout = 30 * time.Second

// WaitUntilRequestProcessed blocks until the request has been processed by the node
func (c *WaspClient) WaitUntilRequestProcessed(chainID isc.ChainID, reqID isc.RequestID, timeout time.Duration) (*isc.Receipt, error) {
	now := time.Now()
	if timeout == 0 {
		timeout = waitRequestProcessedDefaultTimeout
	}
	var res model.RequestReceiptResponse
	err := c.do(
		http.MethodGet,
		routes.WaitRequestProcessed(chainID.String(), reqID.String()),
		&model.WaitRequestProcessedParams{Timeout: timeout},
		&res,
	)
	if err != nil {
		c.log("WaitUntilRequestProcessed, chainID=%v, reqID=%v failed in %v with: %v", chainID, reqID, time.Since(now), err)
		return nil, err
	}
	var receipt isc.Receipt
	err = json.Unmarshal([]byte(res.Receipt), &receipt)
	if err != nil {
		c.log("WaitUntilRequestProcessed, chainID=%v, reqID=%v failed in %v with: %v", chainID, reqID, time.Since(now), err)
		return nil, err
	}
	c.log("WaitUntilRequestProcessed, chainID=%v, reqID=%v done in %v", chainID, reqID, time.Since(now))
	return &receipt, nil
}

// WaitUntilAllRequestsProcessed blocks until all requests in the given transaction have been processed
// by the node
func (c *WaspClient) WaitUntilAllRequestsProcessed(chainID isc.ChainID, tx *iotago.Transaction, timeout time.Duration) ([]*isc.Receipt, error) {
	now := time.Now()
	txID, err := tx.ID()
	if err != nil {
		c.log("WaitUntilAllRequestsProcessed, chainID=%v, tx=? failed in %v with: %v", chainID, time.Since(now), err)
		return nil, err
	}
	reqs, err := isc.RequestsInTransaction(tx)
	if err != nil {
		c.log("WaitUntilAllRequestsProcessed, chainID=%v, tx=%v failed in %v with: %v", chainID, txID.ToHex(), time.Since(now), err)
		return nil, err
	}
	ret := make([]*isc.Receipt, len(reqs))
	for i, req := range reqs[chainID] {
		receipt, err := c.WaitUntilRequestProcessed(chainID, req.ID(), timeout)
		if err != nil {
			c.log("WaitUntilAllRequestsProcessed, chainID=%v, tx=%v failed in %v with: %v", chainID, txID.ToHex(), time.Since(now), err)
			return nil, err
		}
		ret[i] = receipt
	}
	c.log("WaitUntilAllRequestsProcessed, chainID=%v, tx=%v done in %v", chainID, txID.ToHex(), time.Since(now))
	return ret, nil
}
