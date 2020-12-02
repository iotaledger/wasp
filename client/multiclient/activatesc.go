package multiclient

import (
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/coret"
)

func (m *MultiClient) ActivateChain(chainid *coret.ChainID) error {
	return m.Do(func(i int, w *client.WaspClient) error {
		return w.ActivateChain(chainid)
	})
}

func (m *MultiClient) DeactivateChain(chainid *coret.ChainID) error {
	return m.Do(func(i int, w *client.WaspClient) error {
		return w.DeactivateChain(chainid)
	})
}
