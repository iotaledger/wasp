package multiclient

import (
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/registry_pkg/chainrecord"
)

// PutChainRecord calls PutChainRecord in all wasp nodes
func (m *MultiClient) PutChainRecord(bd *chainrecord.ChainRecord) error {
	return m.Do(func(i int, w *client.WaspClient) error {
		return w.PutChainRecord(bd)
	})
}
