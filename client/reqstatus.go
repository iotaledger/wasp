package client

import (
	"net/http"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

// RequestStatus fetches the processing status of a request.
func (c *WaspClient) RequestStatus(chainID *coretypes.ChainID, reqID coretypes.RequestID) (*model.RequestStatusResponse, error) {
	res := &model.RequestStatusResponse{}
	if err := c.do(http.MethodGet, routes.RequestStatus(chainID.Base58(), reqID.Base58()), nil, res); err != nil {
		return nil, err
	}
	return res, nil
}

// WaitUntilRequestProcessed blocks until the request has been processed by the node
func (c *WaspClient) WaitUntilRequestProcessed(chainID *coretypes.ChainID, reqID coretypes.RequestID, timeout time.Duration) error {
	if timeout == 0 {
		timeout = model.WaitRequestProcessedDefaultTimeout
	}
	if err := c.do(
		http.MethodGet,
		routes.WaitRequestProcessed(chainID.Base58(), reqID.Base58()),
		&model.WaitRequestProcessedParams{Timeout: timeout},
		nil,
	); err != nil {
		return err
	}
	return nil
}

// WaitUntilAllRequestsProcessed blocks until all requests in the given transaction have been processed
// by the node
func (c *WaspClient) WaitUntilAllRequestsProcessed(chainID coretypes.ChainID, tx *ledgerstate.Transaction, timeout time.Duration) error {
	for _, out := range tx.Essence().Outputs() {
		if !out.Address().Equals(chainID.AsAddress()) {
			continue
		}
		out, ok := out.(*ledgerstate.ExtendedLockedOutput)
		if !ok {
			continue
		}
		if err := c.WaitUntilRequestProcessed(&chainID, coretypes.RequestID(out.ID()), timeout); err != nil {
			return err
		}
	}
	return nil
}
