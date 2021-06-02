package multiclient

import (
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/coretypes"
)

// WaitUntilRequestProcessed blocks until the request has been processed by all nodes
func (m *MultiClient) WaitUntilRequestProcessed(chainID *coretypes.ChainID, reqID coretypes.RequestID, timeout time.Duration) error {
	oldTimeout := m.Timeout
	defer func() { m.Timeout = oldTimeout }()

	m.Timeout = timeout + 10*time.Second
	return m.Do(func(i int, w *client.WaspClient) error {
		return w.WaitUntilRequestProcessed(chainID, reqID, timeout)
	})
}

// WaitUntilAllRequestsProcessed blocks until all requests in the given transaction have been processed
// by all nodes
func (m *MultiClient) WaitUntilAllRequestsProcessed(chainID coretypes.ChainID, tx *ledgerstate.Transaction, timeout time.Duration) error {
	oldTimeout := m.Timeout
	defer func() { m.Timeout = oldTimeout }()

	m.Timeout = timeout + 10*time.Second
	return m.Do(func(i int, w *client.WaspClient) error {
		return w.WaitUntilAllRequestsProcessed(chainID, tx, timeout)
	})
}
