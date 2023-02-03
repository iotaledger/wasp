package multiclient

import (
	"context"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/packages/registry"
)

// PutChainRecord calls PutChainRecord in all wasp nodes
func (m *MultiClient) PutChainRecord(bd *registry.ChainRecord) error {
	return m.Do(func(i int, w *apiclient.APIClient) error {
		// TODO: Validate the replacement logic from PutChainRecord => ActivateChain + AccessNodes
		_, err := w.ChainsApi.ActivateChain(context.Background(), bd.ChainID().String()).Execute()

		if err != nil {
			return err
		}

		for _, accessNode := range bd.AccessNodes {
			_, err := w.ChainsApi.AddAccessNode(context.Background(), bd.ChainID().String(), accessNode.String()).Execute()

			if err != nil {
				return err
			}
		}

		return nil
	})
}
