package multiclient

import (
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/coretypes"
)

// ActivateChain sends a request to activate a chain in all wasp nodes
func (m *MultiClient) ActivateChain(chainid *coretypes.ChainID) error {
	return m.Do(func(i int, w *client.WaspClient) error {
		return w.ActivateChain(chainid)
	})
}

// DeactivateChain sends a request to deactivate a chain in all wasp nodes
func (m *MultiClient) DeactivateChain(chainid *coretypes.ChainID) error {
	return m.Do(func(i int, w *client.WaspClient) error {
		return w.DeactivateChain(chainid)
	})
}
