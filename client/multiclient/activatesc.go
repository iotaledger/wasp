package multiclient

import (
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/coretypes"
)

func (m *MultiClient) ActivateChain(chainid *coretypes.ChainID) error {
	return m.Do(func(i int, w *client.WaspClient) error {
		return w.ActivateChain(chainid)
	})
}

func (m *MultiClient) DeactivateChain(chainid *coretypes.ChainID) error {
	return m.Do(func(i int, w *client.WaspClient) error {
		return w.DeactivateChain(chainid)
	})
}
