package nodeconn

import (
	"fmt"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/waspconn"
	"github.com/iotaledger/wasp/packages/parameters"
)

func (n *NodeConn) SendWaspIdToNode() {
	n.SendToNode(&waspconn.WaspToNodeSetIdMsg{WaspID: n.netID})
}

func (n *NodeConn) RequestBacklogFromNode(addr *ledgerstate.AliasAddress) {
	n.SendToNode(&waspconn.WaspToNodeGetBacklogMsg{ChainAddress: addr})
}

func (n *NodeConn) RequestConfirmedTransactionFromNode(addr *ledgerstate.AliasAddress, txid ledgerstate.TransactionID) {
	n.SendToNode(&waspconn.WaspToNodeGetConfirmedTransactionMsg{
		ChainAddress: addr,
		TxID:         txid,
	})
}

func (n *NodeConn) RequestTxInclusionStateFromNode(addr *ledgerstate.AliasAddress, txid ledgerstate.TransactionID) {
	n.SendToNode(&waspconn.WaspToNodeGetTxInclusionStateMsg{
		ChainAddress: addr,
		TxID:         txid,
	})
}

func (n *NodeConn) PostTransactionToNode(tx *ledgerstate.Transaction, fromSc ledgerstate.Address, fromLeader uint16) error {
	size := len(tx.Bytes())
	if size > parameters.MaxSerializedTransactionToGoshimmer {
		return fmt.Errorf("size of serialized tx %s is %d --> exceeds maximum size of %d bytes",
			tx.ID(), size, parameters.MaxSerializedTransactionToGoshimmer)
	}
	n.SendToNode(&waspconn.WaspToNodePostTransactionMsg{Tx: tx})
	return nil
}
