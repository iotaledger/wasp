package multiclient

import (
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/iscp"
)

// WaitUntilRequestProcessed blocks until the request has been processed by all nodes
func (m *MultiClient) WaitUntilRequestProcessed(chainID *iscp.ChainID, reqID iscp.RequestID, timeout time.Duration) (*iscp.Receipt, error) {
	oldTimeout := m.Timeout
	defer func() { m.Timeout = oldTimeout }()

	m.Timeout = timeout + 10*time.Second

	var receipt *iscp.Receipt
	var err error
	m.Do(func(i int, w *client.WaspClient) error {
		receipt, err = w.WaitUntilRequestProcessed(chainID, reqID, timeout)
		return err
	})
	return receipt, err
}

// WaitUntilAllRequestsProcessed blocks until all requests in the given transaction have been processed
// by all nodes
func (m *MultiClient) WaitUntilAllRequestsProcessed(chainID *iscp.ChainID, tx *iotago.Transaction, timeout time.Duration) ([]*iscp.Receipt, error) {
	oldTimeout := m.Timeout
	defer func() { m.Timeout = oldTimeout }()

	m.Timeout = timeout + 10*time.Second
	var receipts []*iscp.Receipt
	var err error
	m.Do(func(i int, w *client.WaspClient) error {
		receipts, err = w.WaitUntilAllRequestsProcessed(chainID, tx, timeout)
		return err
	})
	return receipts, err
}
