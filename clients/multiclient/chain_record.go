package multiclient

import (
	"context"

	"github.com/iotaledger/wasp/v2/clients/apiclient"
	"github.com/iotaledger/wasp/v2/packages/registry"
)

// PutChainRecord calls PutChainRecord in all wasp nodes
func (m *MultiClient) PutChainRecord(bd *registry.ChainRecord) error {
	return m.Do(func(i int, w *apiclient.APIClient) error {
		accessNodes := make([]string, len(bd.AccessNodes))

		for k, v := range bd.AccessNodes {
			accessNodes[k] = v.String()
		}

		_, err := w.ChainsAPI.SetChainRecord(context.Background(), bd.ChainID().String()).ChainRecord(apiclient.ChainRecord{
			IsActive:    true,
			AccessNodes: accessNodes,
		}).Execute()

		return err
	})
}
