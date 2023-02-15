package multiclient

import (
	"context"

	"github.com/iotaledger/wasp/clients/apiclient"
)

// ActivateChain sends a request to activate a chain in all wasp nodes
func (m *MultiClient) NodeVersion() (*apiclient.VersionResponse, error) {
	var resp *apiclient.VersionResponse
	err := m.Do(func(i int, w *apiclient.APIClient) error {
		versionResponse, _, err := w.NodeApi.GetVersion(context.Background()).Execute()
		resp = versionResponse
		return err
	})
	return resp, err
}
