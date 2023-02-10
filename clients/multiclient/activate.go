package multiclient

import (
	"context"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/packages/isc"
)

// ActivateChain sends a request to activate a chain in all wasp nodes
func (m *MultiClient) ActivateChain(chainID isc.ChainID) error {
	return m.Do(func(i int, w *apiclient.APIClient) error {
		_, err := w.ChainsApi.ActivateChain(context.Background(), chainID.String()).Execute()
		return err
	})
}

// DeactivateChain sends a request to deactivate a chain in all wasp nodes
func (m *MultiClient) DeactivateChain(chainID isc.ChainID) error {
	return m.Do(func(i int, w *apiclient.APIClient) error {
		_, err := w.ChainsApi.DeactivateChain(context.Background(), chainID.String()).Execute()
		return err
	})
}
