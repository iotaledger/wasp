package multiclient

import (
	"context"

	"github.com/iotaledger/wasp/clients/apiclient"
)

// NodeVersion returns the versions of all the nodes
func (m *MultiClient) NodeVersion() ([]*apiclient.VersionResponse, error) {
	ret := make([]*apiclient.VersionResponse, len(m.nodes))
	err := m.Do(func(i int, w *apiclient.APIClient) error {
		versionResponse, _, err := w.NodeApi.GetVersion(context.Background()).Execute()
		ret[i] = versionResponse
		return err
	})
	return ret, err
}
