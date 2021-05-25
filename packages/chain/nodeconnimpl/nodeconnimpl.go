package nodeconnimpl

import (
	"fmt"

	"github.com/iotaledger/hive.go/logger"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	txstream "github.com/iotaledger/goshimmer/packages/txstream/client"
)

type nodeConnImplementation struct {
	client *txstream.Client
	log    *logger.Logger
}

func New(nodeConnClient *txstream.Client, log *logger.Logger) *nodeConnImplementation {
	return &nodeConnImplementation{
		client: nodeConnClient,
		log:    log,
	}
}

func (n *nodeConnImplementation) PullBacklog(addr *ledgerstate.AliasAddress) {
	fmt.Printf("++++++++++++ pulling backlog %s\n", addr.Base58())
	n.client.RequestBacklog(addr)
}

func (n *nodeConnImplementation) PullState(addr *ledgerstate.AliasAddress) {
	n.client.RequestUnspentAliasOutput(addr)
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

func (n *nodeConnImplementation) PostTransaction(tx *ledgerstate.Transaction) {
	n.client.PostTransaction(tx)
}
