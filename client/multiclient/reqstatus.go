package multiclient

import (
	"fmt"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/isc"
)

// WaitUntilRequestProcessed blocks until the request has been processed by all nodes
func (m *MultiClient) WaitUntilRequestProcessed(chainID isc.ChainID, reqID isc.RequestID, timeout time.Duration) (*isc.Receipt, error) {
	oldTimeout := m.Timeout
	defer func() { m.Timeout = oldTimeout }()

	m.Timeout = timeout

	var receipt *isc.Receipt
	var err error
	err = m.Do(func(i int, w *client.WaspClient) error {
		receipt, err = w.WaitUntilRequestProcessed(chainID, reqID, timeout)
		return err
	})
	return receipt, err
}

// WaitUntilRequestProcessedSuccessfully is similar to WaitUntilRequestProcessed,
// but also checks the receipt and return an error if the request was processed with an error
func (m *MultiClient) WaitUntilRequestProcessedSuccessfully(chainID isc.ChainID, reqID isc.RequestID, timeout time.Duration) (*isc.Receipt, error) {
	receipt, err := m.WaitUntilRequestProcessed(chainID, reqID, timeout)
	if err != nil {
		return receipt, err
	}
	if receipt.ResolvedError != "" {
		return receipt, fmt.Errorf("request processed with an error: %s", receipt.ResolvedError)
	}
	return receipt, nil
}

// WaitUntilAllRequestsProcessed blocks until all requests in the given transaction have been processed
// by all nodes
func (m *MultiClient) WaitUntilAllRequestsProcessed(chainID isc.ChainID, tx *iotago.Transaction, timeout time.Duration) ([]*isc.Receipt, error) {
	oldTimeout := m.Timeout
	defer func() { m.Timeout = oldTimeout }()

	m.Timeout = timeout
	var receipts []*isc.Receipt
	var err error
	err = m.Do(func(i int, w *client.WaspClient) error {
		receipts, err = w.WaitUntilAllRequestsProcessed(chainID, tx, timeout)
		return err
	})
	return receipts, err
}

// WaitUntilAllRequestsProcessedSuccessfully is similar to WaitUntilAllRequestsProcessed
// but also checks the receipts and return an error if any of the requests was processed with an error
func (m *MultiClient) WaitUntilAllRequestsProcessedSuccessfully(chainID isc.ChainID, tx *iotago.Transaction, timeout time.Duration) ([]*isc.Receipt, error) {
	receipts, err := m.WaitUntilAllRequestsProcessed(chainID, tx, timeout)
	if err != nil {
		return receipts, err
	}
	for i, receipt := range receipts {
		if receipt.ResolvedError != "" {
			return receipts, fmt.Errorf("error found on receipt #%d: %s", i, receipt.ResolvedError)
		}
	}
	return receipts, nil
}
