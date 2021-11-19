package nodeconnimpl

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/iscp"
)

type ChainNodeConnImplementation struct {
	client chain.NodeConnection
	log    *logger.Logger // each chain has its own logger
}

var _ chain.NodeConnectionSender = &ChainNodeConnImplementation{}

func NewChainNodeConnImplementation(nodeConnClient chain.NodeConnection, log *logger.Logger) chain.NodeConnectionSender {
	return &ChainNodeConnImplementation{
		client: nodeConnClient,
		log:    log,
	}
}

func (n *ChainNodeConnImplementation) PullBacklog(addr *ledgerstate.AliasAddress) {
	n.log.Debugf("ChainNodeConnImplementation::PullBacklog(addr=%s)...", addr.Base58())
	n.client.PullBacklog(addr)
	n.log.Debugf("ChainNodeConnImplementation::PullBacklog(addr=%s)... Done", addr.Base58())
}

func (n *ChainNodeConnImplementation) PullState(addr *ledgerstate.AliasAddress) {
	n.log.Debugf("ChainNodeConnImplementation::PullState(addr=%s)...", addr.Base58())
	n.client.PullState(addr)
	n.log.Debugf("ChainNodeConnImplementation::PullState(addr=%s)... Done", addr.Base58())
}

func (n *ChainNodeConnImplementation) PullConfirmedTransaction(addr ledgerstate.Address, txID ledgerstate.TransactionID) {
	n.log.Debugf("ChainNodeConnImplementation::PullConfirmedTransaction(addr=%s, txID=%v)...", addr.Base58(), txID.Base58())
	n.client.PullConfirmedTransaction(addr, txID)
	n.log.Debugf("ChainNodeConnImplementation::PullConfirmedTransaction(addr=%s, txID=%v)... Done", addr.Base58(), txID.Base58())
}

func (n *ChainNodeConnImplementation) PullTransactionInclusionState(addr ledgerstate.Address, txID ledgerstate.TransactionID) {
	n.log.Debugf("ChainNodeConnImplementation::PullTransactionInclusionState(addr=%s, txID=%v)...", addr.Base58(), txID.Base58())
	n.client.PullTransactionInclusionState(addr, txID)
	n.log.Debugf("ChainNodeConnImplementation::PullTransactionInclusionState(addr=%s, txID=%v)... Done", addr.Base58(), txID.Base58())
}

func (n *ChainNodeConnImplementation) PullConfirmedOutput(addr ledgerstate.Address, outputID ledgerstate.OutputID) {
	outputIDStr := iscp.OID(outputID)
	n.log.Debugf("ChainNodeConnImplementation::PullConfirmedOutput(addr=%s, outputID=%v)...", addr.Base58(), outputIDStr)
	n.client.PullConfirmedOutput(addr, outputID)
	n.log.Debugf("ChainNodeConnImplementation::PullConfirmedOutput(addr=%s, outputID=%v)... Done", addr.Base58(), outputIDStr)
}

func (n *ChainNodeConnImplementation) PostTransaction(tx *ledgerstate.Transaction) {
	n.log.Debugf("ChainNodeConnImplementation::PostTransaction(txID=%s)...", tx.ID().Base58())
	n.client.PostTransaction(tx)
	n.log.Debugf("ChainNodeConnImplementation::PostTransaction(txID=%s)... Done", tx.ID().Base58())
}
