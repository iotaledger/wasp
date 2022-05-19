package multiclient

import (
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/iscp"
	"golang.org/x/xerrors"
)

// WaitUntilRequestProcessed blocks until the request has been processed by all nodes
func (m *MultiClient) WaitUntilRequestProcessed(chainID *iscp.ChainID, reqID iscp.RequestID, timeout time.Duration) (*iscp.Receipt, error) {
	oldTimeout := m.Timeout
	defer func() { m.Timeout = oldTimeout }()

	m.Timeout = timeout + 10*time.Second

	var receipt *iscp.Receipt
	var err error
	err = m.Do(func(i int, w *client.WaspClient) error {
		receipt, err = w.WaitUntilRequestProcessed(chainID, reqID, timeout)
		return err
	})
	return receipt, err
}

// WaitUntilRequestProcessedSuccessfully is similar to WaitUntilRequestProcessed,
// but also checks the receipt and return an error if the request was processed with an error
func (m *MultiClient) WaitUntilRequestProcessedSuccessfully(chainID *iscp.ChainID, reqID iscp.RequestID, timeout time.Duration) (*iscp.Receipt, error) {
	receipt, err := m.WaitUntilRequestProcessed(chainID, reqID, timeout)
	if err != nil {
		return receipt, err
	}
	if receipt.TranslatedError != "" {
		return receipt, xerrors.Errorf("request processed with an error: %s", receipt.TranslatedError)
	}
	return receipt, nil
}

// WaitUntilAllRequestsProcessed blocks until all requests in the given transaction have been processed
// by all nodes
func (m *MultiClient) WaitUntilAllRequestsProcessed(chainID *iscp.ChainID, tx *iotago.Transaction, timeout time.Duration) ([]*iscp.Receipt, error) {
	oldTimeout := m.Timeout
	defer func() { m.Timeout = oldTimeout }()

	m.Timeout = timeout + 10*time.Second
	var receipts []*iscp.Receipt
	var err error
	err = m.Do(func(i int, w *client.WaspClient) error {
		receipts, err = w.WaitUntilAllRequestsProcessed(chainID, tx, timeout)
		return err
	})
	return receipts, err
}

// WaitUntilAllRequestsProcessedSuccessfully is similar to WaitUntilAllRequestsProcessed
// but also checks the receipts and return an error if any of the requests was processed with an error
func (m *MultiClient) WaitUntilAllRequestsProcessedSuccessfully(chainID *iscp.ChainID, tx *iotago.Transaction, timeout time.Duration) ([]*iscp.Receipt, error) {
	receipts, err := m.WaitUntilAllRequestsProcessed(chainID, tx, timeout)
	if err != nil {
		return receipts, err
	}
	for i, receipt := range receipts {
		if receipt.TranslatedError != "" {
			return receipts, xerrors.Errorf("error found on receipt #%d: %s", i, receipt.TranslatedError)
		}
	}
	return receipts, nil
}
