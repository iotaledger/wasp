package nodeconnimpl

import (
	"github.com/iotaledger/hive.go/logger"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	txstream "github.com/iotaledger/goshimmer/packages/txstream/client"
)

type NodeConnImplementation struct {
	client *txstream.Client
	log    *logger.Logger
}

func New(nodeConnClient *txstream.Client, log *logger.Logger) *NodeConnImplementation {
	return &NodeConnImplementation{
		client: nodeConnClient,
		log:    log,
	}
}

func (n *NodeConnImplementation) PullBacklog(addr *ledgerstate.AliasAddress) {
	n.log.Debugf("NodeConnImplementation::PullBacklog(addr=%s)...", addr.Base58())
	n.client.RequestBacklog(addr)
	n.log.Debugf("NodeConnImplementation::PullBacklog(addr=%s)... Done", addr.Base58())
}

func (n *NodeConnImplementation) PullState(addr *ledgerstate.AliasAddress) {
	n.log.Debugf("NodeConnImplementation::PullState(addr=%s)...", addr.Base58())
	n.client.RequestUnspentAliasOutput(addr)
	n.log.Debugf("NodeConnImplementation::PullState(addr=%s)... Done", addr.Base58())
}

func (n *NodeConnImplementation) PullConfirmedTransaction(addr ledgerstate.Address, txid ledgerstate.TransactionID) {
	n.log.Debugf("NodeConnImplementation::PullConfirmedTransaction(addr=%s)...", addr.Base58())
	n.client.RequestConfirmedTransaction(addr, txid)
	n.log.Debugf("NodeConnImplementation::PullConfirmedTransaction(addr=%s)... Done", addr.Base58())
}

func (n *NodeConnImplementation) PullTransactionInclusionState(addr ledgerstate.Address, txid ledgerstate.TransactionID) {
	n.log.Debugf("NodeConnImplementation::PullTransactionInclusionState(addr=%s)...", addr.Base58())
	n.client.RequestTxInclusionState(addr, txid)
	n.log.Debugf("NodeConnImplementation::PullTransactionInclusionState(addr=%s)... Done", addr.Base58())
}

func (n *NodeConnImplementation) PullConfirmedOutput(addr ledgerstate.Address, outputID ledgerstate.OutputID) {
	n.log.Debugf("NodeConnImplementation::PullConfirmedOutput(addr=%s)...", addr.Base58())
	n.client.RequestConfirmedOutput(addr, outputID)
	n.log.Debugf("NodeConnImplementation::PullConfirmedOutput(addr=%s)... Done", addr.Base58())
}

func (n *NodeConnImplementation) PostTransaction(tx *ledgerstate.Transaction) {
	n.log.Debugf("NodeConnImplementation::PostTransaction(tx, id=%s)...", tx.ID())
	n.client.PostTransaction(tx)
	n.log.Debugf("NodeConnImplementation::PostTransaction(tx, id=%s)... Done", tx.ID())
}
