package multiclient

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/client"
)

// DKSharesGet retrieves distributed key info with specific ChainID from multiple hosts.
func (m *MultiClient) DKSharesGet(sharedAddress *address.Address) ([]*client.DKSharesInfo, error) {
	ret := make([]*client.DKSharesInfo, len(m.nodes))
	err := m.Do(func(i int, w *client.WaspClient) error {
		k, err := w.DKSharesGet(sharedAddress)
		ret[i] = k
		return err
	})
	return ret, err
}
