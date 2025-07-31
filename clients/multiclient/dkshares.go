package multiclient

import (
	"context"

	"github.com/iotaledger/wasp/v2/clients/apiclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
)

// DKSharesGet retrieves distributed key info with specific ChainID from multiple hosts.
func (m *MultiClient) DKSharesGet(sharedAddress iotago.Address) ([]*apiclient.DKSharesInfo, error) {
	ret := make([]*apiclient.DKSharesInfo, len(m.nodes))
	err := m.Do(func(i int, w *apiclient.APIClient) error {
		sharesInfo, _, err := w.NodeAPI.GetDKSInfo(context.Background(), sharedAddress.String()).Execute()
		ret[i] = sharesInfo
		return err
	})
	return ret, err
}
