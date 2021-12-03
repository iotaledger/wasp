package multiclient

import (
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/iscp"
)

// WaitUntilRequestProcessed blocks until the request has been processed by all nodes
func (m *MultiClient) WaitUntilRequestProcessed(chainID *iscp.ChainID, reqID iscp.RequestID, timeout time.Duration) error {
	oldTimeout := m.Timeout
	defer func() { m.Timeout = oldTimeout }()

	m.Timeout = timeout + 10*time.Second
	return m.Do(func(i int, w *client.WaspClient) error {
		return w.WaitUntilRequestProcessed(chainID, reqID, timeout)
	})
}

// WaitUntilAllRequestsProcessed blocks until all requests in the given transaction have been processed
// by all nodes
func (m *MultiClient) WaitUntilAllRequestsProcessed(chainID *iscp.ChainID, tx *iotago.Transaction, timeout time.Duration) error {
	oldTimeout := m.Timeout
	defer func() { m.Timeout = oldTimeout }()

	m.Timeout = timeout + 10*time.Second
	return m.Do(func(i int, w *client.WaspClient) error {
		return w.WaitUntilAllRequestsProcessed(chainID, tx, timeout)
	})
}
