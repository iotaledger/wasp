package chains

import (
	"github.com/iotaledger/goshimmer/packages/waspconn"
	"github.com/iotaledger/wasp/packages/coretypes"
)

func processNodeMsg(msg interface{}) {
	switch msgt := msg.(type) {
	case *waspconn.WaspFromNodeTransactionMsg:
		chainID := coretypes.NewChainID(msgt.ChainAddress)
		chain := GetChain(chainID)
		if chain == nil {
			return
		}
		log.Debugw("dispatch transaction",
			"txid", msgt.Tx.ID().String(),
			"chainid", chainID.String(),
		)
		chain.ReceiveMessage(msgt)

	case *waspconn.WaspFromNodeTxInclusionStateMsg:
		chainID := coretypes.NewChainID(msgt.ChainAddress)
		ch := GetChain(chainID)
		if ch == nil {
			return
		}
		ch.ReceiveMessage(msgt)
	}
	log.Errorf("wrong message type")
}
