package multiclient

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/webapi/model"
)

// DKSharesGet retrieves distributed key info with specific ChainID from multiple hosts.
func (m *MultiClient) DKSharesGet(sharedAddress ledgerstate.Address) ([]*model.DKSharesInfo, error) {
	ret := make([]*model.DKSharesInfo, len(m.nodes))
	err := m.Do(func(i int, w *client.WaspClient) error {
		k, err := w.DKSharesGet(sharedAddress)
		ret[i] = k
		return err
	})
	return ret, err
}
