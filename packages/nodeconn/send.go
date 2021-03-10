package nodeconn

import (
	"fmt"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/waspconn"
	"github.com/iotaledger/wasp/packages/parameters"
)

func (n *NodeConn) SendWaspIdToNode() error {
	data := waspconn.EncodeMsg(&waspconn.WaspToNodeSetIdMsg{
		Waspid: n.netID,
	})
	if err := n.SendDataToNode(data); err != nil {
		return err
	}
	return nil
}

func (n *NodeConn) RequestOutputsFromNode(addr ledgerstate.Address) error {
	data := waspconn.EncodeMsg(&waspconn.WaspToNodeGetOutputsMsg{
		Address: addr,
	})
	if err := n.SendDataToNode(data); err != nil {
		return err
	}
	return nil
}

func (n *NodeConn) RequestConfirmedTransactionFromNode(txid ledgerstate.TransactionID) error {
	data := waspconn.EncodeMsg(&waspconn.WaspToNodeGetConfirmedTransactionMsg{
		TxId: txid,
	})
	if err := n.SendDataToNode(data); err != nil {
		return err
	}
	return nil
}

func (n *NodeConn) RequestBranchInclusionStateFromNode(txid ledgerstate.TransactionID, addr ledgerstate.Address) error {
	n.log.Debugf("RequestInclusionLevelFromNode. txid %s", txid.String())

	data := waspconn.EncodeMsg(&waspconn.WaspToNodeGetBranchInclusionStateMsg{
		TxId:      txid,
		SCAddress: addr,
	})
	if err := n.SendDataToNode(data); err != nil {
		return err
	}
	return nil

}

func (n *NodeConn) PostTransactionToNode(tx *ledgerstate.Transaction, fromSc ledgerstate.Address, fromLeader uint16) error {
	data := waspconn.EncodeMsg(&waspconn.WaspToNodeTransactionMsg{
		Tx:        tx,
		SCAddress: fromSc, // just for tracing
		Leader:    fromLeader,
	})
	if len(data) > parameters.MaxSerializedTransactionToGoshimmer {
		return fmt.Errorf("size if serialized tx %s is %d --> exceeds maximum size of %d bytes",
			tx.ID(), len(data), parameters.MaxSerializedTransactionToGoshimmer)
	}
	if err := n.SendDataToNode(data); err != nil {
		return err
	}
	return nil
}
