package multiclient

import (
	"time"

	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

func (m *MultiClient) WaitUntilRequestProcessed(chainId *coretypes.ChainID, reqId *coretypes.RequestID, timeout time.Duration) error {
	return m.Do(func(i int, w *client.WaspClient) error {
		return w.WaitUntilRequestProcessed(chainId, reqId, timeout)
	})
}

func (m *MultiClient) WaitUntilAllRequestsProcessed(tx *sctransaction.Transaction, timeout time.Duration) error {
	return m.Do(func(i int, w *client.WaspClient) error {
		return w.WaitUntilAllRequestsProcessed(tx, timeout)
	})
}
