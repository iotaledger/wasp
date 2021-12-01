// Provides implementations for chain.ChainNodeConnection methods
package nodeconnimpl

import (
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
)

type chainNodeConnImplementation struct {
	chainID  *iscp.ChainID
	nodeConn chain.NodeConnection
	metrics  nodeconnmetrics.NodeConnectionMessagesMetrics
	log      *logger.Logger // logger for one chain
}

var _ chain.ChainNodeConnection = &chainNodeConnImplementation{}

func NewChainNodeConnection(chainID *iscp.ChainID, nodeConn chain.NodeConnection, log *logger.Logger) chain.ChainNodeConnection {
	return &chainNodeConnImplementation{
		chainID:  chainID,
		nodeConn: nodeConn,
		metrics:  nodeConn.GetMetrics().NewMessagesMetrics(chainID),
		log:      log,
	}
}

func (c *chainNodeConnImplementation) AttachToTransactionReceived(fun chain.NodeConnectionHandleTransactionFun) {
	c.log.Debugf("ChainNodeConnImplementation::AttachToTransactionReceived")
	c.nodeConn.AttachToTransactionReceived(c.chainID.AsAliasAddress(), func(tx *ledgerstate.Transaction) {
		c.metrics.GetInTransaction().CountLastMessage(tx)
		fun(tx)
	})
}

func (c *chainNodeConnImplementation) AttachToInclusionStateReceived(fun chain.NodeConnectionHandleInclusionStateFun) {
	c.log.Debugf("ChainNodeConnImplementation::AttachToInclusionStateReceived")
	c.nodeConn.AttachToInclusionStateReceived(c.chainID.AsAliasAddress(), func(txID ledgerstate.TransactionID, iState ledgerstate.InclusionState) {
		c.metrics.GetInInclusionState().CountLastMessage(struct {
			TransactionID  ledgerstate.TransactionID
			InclusionState ledgerstate.InclusionState
		}{
			TransactionID:  txID,
			InclusionState: iState,
		})
		fun(txID, iState)
	})
}

func (c *chainNodeConnImplementation) AttachToOutputReceived(fun chain.NodeConnectionHandleOutputFun) {
	c.log.Debugf("ChainNodeConnImplementation::AttachToOutputReceived")
	c.nodeConn.AttachToOutputReceived(c.chainID.AsAliasAddress(), func(output ledgerstate.Output) {
		c.metrics.GetInOutput().CountLastMessage(output)
		fun(output)
	})
}

func (c *chainNodeConnImplementation) AttachToUnspentAliasOutputReceived(fun chain.NodeConnectionHandleUnspentAliasOutputFun) {
	c.log.Debugf("ChainNodeConnImplementation::AttachToUnspentAliasOutputReceived")
	c.nodeConn.AttachToUnspentAliasOutputReceived(c.chainID.AsAliasAddress(), func(output *ledgerstate.AliasOutput, timestamp time.Time) {
		c.metrics.GetInUnspentAliasOutput().CountLastMessage(struct {
			AliasOutput *ledgerstate.AliasOutput
			Timestamp   time.Time
		}{
			AliasOutput: output,
			Timestamp:   timestamp,
		})
		fun(output, timestamp)
	})
}

func (c *chainNodeConnImplementation) DetachFromTransactionReceived() {
	c.log.Debugf("ChainNodeConnImplementation::DetachFromTransactionReceived")
	c.nodeConn.DetachFromTransactionReceived(c.chainID.AsAliasAddress())
}

func (c *chainNodeConnImplementation) DetachFromInclusionStateReceived() {
	c.log.Debugf("ChainNodeConnImplementation::DetachFromInclusionStateReceived")
	c.nodeConn.DetachFromInclusionStateReceived(c.chainID.AsAliasAddress())
}

func (c *chainNodeConnImplementation) DetachFromOutputReceived() {
	c.log.Debugf("ChainNodeConnImplementation::DetachFromOutputReceived")
	c.nodeConn.DetachFromOutputReceived(c.chainID.AsAliasAddress())
}

func (c *chainNodeConnImplementation) DetachFromUnspentAliasOutputReceived() {
	c.log.Debugf("ChainNodeConnImplementation::DetachFromUnspentAliasOutputReceived")
	c.nodeConn.DetachFromUnspentAliasOutputReceived(c.chainID.AsAliasAddress())
}

func (c *chainNodeConnImplementation) PullState() {
	c.log.Debugf("ChainNodeConnection::PullState...")
	c.metrics.GetOutPullState().CountLastMessage(nil)
	c.nodeConn.PullState(c.chainID.AsAliasAddress())
	c.log.Debugf("ChainNodeConnection::PullState... Done")
}

func (c *chainNodeConnImplementation) PullTransactionInclusionState(txID ledgerstate.TransactionID) {
	txIDStr := txID.Base58()
	c.log.Debugf("ChainNodeConnImplementation::PullTransactionInclusionState(txID=%v)...", txIDStr)
	c.metrics.GetOutPullTransactionInclusionState().CountLastMessage(txID)
	c.nodeConn.PullTransactionInclusionState(c.chainID.AsAddress(), txID)
	c.log.Debugf("ChainNodeConnImplementation::PullTransactionInclusionState(txID=%v)... Done", txIDStr)
}

func (c *chainNodeConnImplementation) PullConfirmedOutput(outputID ledgerstate.OutputID) {
	outputIDStr := iscp.OID(outputID)
	c.log.Debugf("ChainNodeConnImplementation::PullConfirmedOutput(outputID=%v)...", outputIDStr)
	c.metrics.GetOutPullConfirmedOutput().CountLastMessage(outputID)
	c.nodeConn.PullConfirmedOutput(c.chainID.AsAddress(), outputID)
	c.log.Debugf("ChainNodeConnImplementation::PullConfirmedOutput(outputID=%v)... Done", outputIDStr)
}

func (c *chainNodeConnImplementation) PostTransaction(tx *ledgerstate.Transaction) {
	txIDStr := tx.ID().Base58()
	c.log.Debugf("ChainNodeConnImplementation::PostTransaction(txID=%s)...", txIDStr)
	c.metrics.GetOutPostTransaction().CountLastMessage(tx)
	c.nodeConn.PostTransaction(tx)
	c.log.Debugf("ChainNodeConnImplementation::PostTransaction(txID=%s)... Done", txIDStr)
}

func (c *chainNodeConnImplementation) GetMetrics() nodeconnmetrics.NodeConnectionMessagesMetrics {
	return c.metrics
}

func (c *chainNodeConnImplementation) Close() {
	c.log.Debugf("ChainNodeConnImplementation::Close")
	c.DetachFromTransactionReceived()
	c.DetachFromInclusionStateReceived()
	c.DetachFromOutputReceived()
	c.DetachFromUnspentAliasOutputReceived()
}
