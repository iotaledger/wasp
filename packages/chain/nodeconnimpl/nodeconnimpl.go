package nodeconnimpl

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	txstream "github.com/iotaledger/goshimmer/packages/txstream/client"
)

type nodeConnImplementation struct {
	client *txstream.Client
}

func New(nodeConnClient *txstream.Client) *nodeConnImplementation {
	return &nodeConnImplementation{client: nodeConnClient}
}

func (n *nodeConnImplementation) PullBacklog(addr ledgerstate.Address) {
	n.client.RequestBacklog(addr)
}

func (n *nodeConnImplementation) PullState(addr ledgerstate.Address) {
	// TODO adjust nodeconn
	n.client.RequestBacklog(addr)
}

func (n *nodeConnImplementation) PullConfirmedTransaction(addr ledgerstate.Address, txid ledgerstate.TransactionID) {
	n.client.RequestConfirmedTransaction(addr, txid)
}

func (n *nodeConnImplementation) PullTransactionInclusionState(addr ledgerstate.Address, txid ledgerstate.TransactionID) {
	n.client.RequestTxInclusionState(addr, txid)
}

func (n *nodeConnImplementation) PullConfirmedOutput(addr ledgerstate.Address, outputID ledgerstate.OutputID) {
	n.client.RequestConfirmedOutput(addr, outputID)
}

func (n *nodeConnImplementation) PostTransaction(tx *ledgerstate.Transaction, from ledgerstate.Address, fromLeader uint16) {
	n.client.PostTransaction(tx, from, fromLeader)
}
