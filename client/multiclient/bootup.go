package multiclient

import (
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/registry"
)

// PutBootupData calls PutBootupData to hosts in parallel
func (m *MultiClient) PutBootupData(bd *registry.BootupData) error {
	return m.Do(func(i int, w *client.WaspClient) error {
		return w.PutBootupData(bd)
	})
}
