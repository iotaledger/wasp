package multiclient

import (
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/registry"
)

// PutChainRecord calls PutChainRecord to hosts in parallel
func (m *MultiClient) PutChainRecord(bd *registry.ChainRecord) error {
	return m.Do(func(i int, w *client.WaspClient) error {
		return w.PutChainRecord(bd)
	})
}
