package multiclient

import (
	"time"

	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/sctransaction_old"
)

// WaitUntilRequestProcessed blocks until the request has been processed by all nodes
func (m *MultiClient) WaitUntilRequestProcessed(chainId *coretypes.ChainID, reqId *coretypes.RequestID, timeout time.Duration) error {
	oldTimeout := m.Timeout
	defer func() { m.Timeout = oldTimeout }()

	m.Timeout = timeout + 10*time.Second
	return m.Do(func(i int, w *client.WaspClient) error {
		return w.WaitUntilRequestProcessed(chainId, reqId, timeout)
	})
}

// WaitUntilAllRequestsProcessed blocks until all requests in the given transaction have been processed
// by all nodes
func (m *MultiClient) WaitUntilAllRequestsProcessed(tx *sctransaction_old.TransactionEssence, timeout time.Duration) error {
	oldTimeout := m.Timeout
	defer func() { m.Timeout = oldTimeout }()

	m.Timeout = timeout + 10*time.Second
	return m.Do(func(i int, w *client.WaspClient) error {
		return w.WaitUntilAllRequestsProcessed(tx, timeout)
	})
}
