// Package multiclient provides functionality for managing multiple client connections.
package multiclient

import (
	"context"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/packages/isc"
)

// ActivateChain sends a request to activate a chain in all wasp nodes
func (m *MultiClient) ActivateChain(chainID isc.ChainID) error {
	return m.Do(func(i int, w *apiclient.APIClient) error {
		_, err := w.ChainsAPI.ActivateChain(context.Background(), chainID.String()).Execute()
		return err
	})
}

// DeactivateChain sends a request to deactivate a chain in all wasp nodes
func (m *MultiClient) DeactivateChain() error {
	return m.Do(func(i int, w *apiclient.APIClient) error {
		_, err := w.ChainsAPI.DeactivateChain(context.Background()).Execute()
		return err
	})
}
