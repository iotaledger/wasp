package multiclient

import (
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/isc"
)

// ActivateChain sends a request to activate a chain in all wasp nodes
func (m *MultiClient) ActivateChain(chID *isc.ChainID) error {
	return m.Do(func(i int, w *client.WaspClient) error {
		return w.ActivateChain(chID)
	})
}

// DeactivateChain sends a request to deactivate a chain in all wasp nodes
func (m *MultiClient) DeactivateChain(chID *isc.ChainID) error {
	return m.Do(func(i int, w *client.WaspClient) error {
		return w.DeactivateChain(chID)
	})
}
