package multiclient

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/client"
)

// GetPublicKeyInfo retrieves public info about key with specific address from multiple hosts
func (m *MultiClient) GetPublicKeyInfo(addr *address.Address) ([]*client.PubKeyInfo, error) {
	ret := make([]*client.PubKeyInfo, len(m.nodes))
	err := m.Do(func(i int, w *client.WaspClient) error {
		k, err := w.GetPublicKeyInfo(addr)
		ret[i] = k
		return err
	})
	return ret, err
}
