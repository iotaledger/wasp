package nodeconnimpl

import (
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/txstream"
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

func (n *NodeConnImplementation) PullBacklog(addr *iotago.AliasAddress) {
	n.log.Debugf("NodeConnImplementation::PullBacklog(addr=%s)...", addr.Bech32(iscp.Bech32Prefix))
	n.client.RequestBacklog(addr)
	n.log.Debugf("NodeConnImplementation::PullBacklog(addr=%s)... Done", addr.Bech32(iscp.Bech32Prefix))
}

func (n *NodeConnImplementation) PullState(addr *iotago.AliasAddress) {
	n.log.Debugf("NodeConnImplementation::PullState(addr=%s)...", addr.Bech32(iscp.Bech32Prefix))
	n.client.RequestUnspentAliasOutput(addr)
	n.log.Debugf("NodeConnImplementation::PullState(addr=%s)... Done", addr.Bech32(iscp.Bech32Prefix))
}

func (n *NodeConnImplementation) PullConfirmedTransaction(addr iotago.Address, txid iotago.TransactionID) {
	n.log.Debugf("NodeConnImplementation::PullConfirmedTransaction(addr=%s)...", addr.Bech32(iscp.Bech32Prefix))
	n.client.RequestConfirmedTransaction(addr, txid)
	n.log.Debugf("NodeConnImplementation::PullConfirmedTransaction(addr=%s)... Done", addr.Bech32(iscp.Bech32Prefix))
}

func (n *NodeConnImplementation) PullTransactionInclusionState(addr iotago.Address, txid iotago.TransactionID) {
	n.log.Debugf("NodeConnImplementation::PullTransactionInclusionState(addr=%s)...", addr.Bech32(iscp.Bech32Prefix))
	n.client.RequestTxInclusionState(addr, txid)
	n.log.Debugf("NodeConnImplementation::PullTransactionInclusionState(addr=%s)... Done", addr.Bech32(iscp.Bech32Prefix))
}

func (n *NodeConnImplementation) PullConfirmedOutput(addr iotago.Address, outputID iotago.OutputID) {
	n.log.Debugf("NodeConnImplementation::PullConfirmedOutput(addr=%s)...", addr.Bech32(iscp.Bech32Prefix))
	n.client.RequestConfirmedOutput(addr, outputID)
	n.log.Debugf("NodeConnImplementation::PullConfirmedOutput(addr=%s)... Done", addr.Bech32(iscp.Bech32Prefix))
}

func (n *NodeConnImplementation) PostTransaction(tx *iotago.Transaction) {
	n.log.Debugf("NodeConnImplementation::PostTransaction(tx, id=%s)...", tx.ID())
	n.client.PostTransaction(tx)
	n.log.Debugf("NodeConnImplementation::PostTransaction(tx, id=%s)... Done", tx.ID())
}
