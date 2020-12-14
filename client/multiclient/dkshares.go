package multiclient

import (
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/coretypes"
)

// DKSharesGet retrieves distributed key info with specific ChainID from multiple hosts.
func (m *MultiClient) DKSharesGet(chainID *coretypes.ChainID) ([]*client.DKSharesInfo, error) {
	ret := make([]*client.DKSharesInfo, len(m.nodes))
	err := m.Do(func(i int, w *client.WaspClient) error {
		k, err := w.DKSharesGet(chainID)
		ret[i] = k
		return err
	})
	return ret, err
}
