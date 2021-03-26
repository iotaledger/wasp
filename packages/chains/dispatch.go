package chains

import (
	"github.com/iotaledger/wasp/packages/coretypes"
)

func (c *Chains) dispatchMsg(msg interface{}) {
	switch msgt := msg.(type) {
	case *waspconn.WaspFromNodeTransactionMsg:
		chainID := coretypes.NewChainID(msgt.ChainAddress)
		chain := c.Get(chainID)
		if chain == nil {
			return
		}
		c.log.Debugw("dispatch transaction",
			"txid", msgt.Tx.ID().String(),
			"chainid", chainID.String(),
		)
		chain.ReceiveMessage(msgt)

	case *waspconn.WaspFromNodeTxInclusionStateMsg:
		chainID := coretypes.NewChainID(msgt.ChainAddress)
		ch := c.Get(chainID)
		if ch == nil {
			return
		}
		ch.ReceiveMessage(msgt)
	}
	c.log.Errorf("wrong message type")
}
