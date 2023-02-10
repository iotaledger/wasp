package multiclient

import (
	"context"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/apiclient"
)

// DKSharesGet retrieves distributed key info with specific ChainID from multiple hosts.
func (m *MultiClient) DKSharesGet(sharedAddress iotago.Address) ([]*apiclient.DKSharesInfo, error) {
	ret := make([]*apiclient.DKSharesInfo, len(m.nodes))
	err := m.Do(func(i int, w *apiclient.APIClient) error {
		sharesInfo, _, err := w.NodeApi.GetDKSInfo(context.Background(), sharedAddress.String()).Execute()
		ret[i] = sharesInfo
		return err
	})
	return ret, err
}
