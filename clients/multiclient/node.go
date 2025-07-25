package multiclient

import (
	"context"

	"github.com/iotaledger/wasp/v2/clients/apiclient"
)

// NodeVersion returns the versions of all the nodes
func (m *MultiClient) NodeVersion() ([]*apiclient.VersionResponse, error) {
	ret := make([]*apiclient.VersionResponse, len(m.nodes))
	err := m.Do(func(i int, w *apiclient.APIClient) error {
		versionResponse, _, err := w.NodeAPI.GetVersion(context.Background()).Execute()
		ret[i] = versionResponse
		return err
	})
	return ret, err
}
