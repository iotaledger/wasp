package nodeconn

import (
	"fmt"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/waspconn"
	"github.com/iotaledger/wasp/packages/parameters"
)

func (n *NodeConn) SendWaspIdToNode() error {
	return n.sendToNode(&waspconn.WaspToNodeSetIdMsg{WaspID: n.netID})
}

func (n *NodeConn) RequestBacklogFromNode(addr *ledgerstate.AliasAddress) error {
	return n.sendToNode(&waspconn.WaspToNodeGetBacklogMsg{ChainAddress: addr})
}

func (n *NodeConn) RequestConfirmedTransactionFromNode(addr *ledgerstate.AliasAddress, txid ledgerstate.TransactionID) error {
	return n.sendToNode(&waspconn.WaspToNodeGetConfirmedTransactionMsg{
		ChainAddress: addr,
		TxID:         txid,
	})
}

func (n *NodeConn) RequestTxInclusionStateFromNode(addr *ledgerstate.AliasAddress, txid ledgerstate.TransactionID) error {
	return n.sendToNode(&waspconn.WaspToNodeGetTxInclusionStateMsg{
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
	return n.sendToNode(&waspconn.WaspToNodePostTransactionMsg{Tx: tx})
}
