// Provides implementations for chain.ChainNodeConnection methods
package chainimpl

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/iscp"
)

func (c *chainObj) AttachToTransactionReceived(fun chain.NodeConnectionHandleTransactionFun) {
	c.nodeConn.AttachToTransactionReceived(c.ID().AsAliasAddress(), fun)
}

func (c *chainObj) AttachToInclusionStateReceived(fun chain.NodeConnectionHandleInclusionStateFun) {
	c.nodeConn.AttachToInclusionStateReceived(c.ID().AsAliasAddress(), fun)
}

func (c *chainObj) AttachToOutputReceived(fun chain.NodeConnectionHandleOutputFun) {
	c.nodeConn.AttachToOutputReceived(c.ID().AsAliasAddress(), fun)
}

func (c *chainObj) AttachToUnspentAliasOutputReceived(fun chain.NodeConnectionHandleUnspentAliasOutputFun) {
	c.nodeConn.AttachToUnspentAliasOutputReceived(c.ID().AsAliasAddress(), fun)
}

func (c *chainObj) PullState() {
	c.log.Debugf("ChainNodeConnection::PullState...")
	c.nodeConn.PullState(c.ID().AsAliasAddress())
	c.log.Debugf("ChainNodeConnection::PullState... Done")
}

func (c *chainObj) PullTransactionInclusionState(txID ledgerstate.TransactionID) {
	txIDStr := txID.Base58()
	c.log.Debugf("ChainNodeConnImplementation::PullTransactionInclusionState(txID=%v)...", txIDStr)
	c.nodeConn.PullTransactionInclusionState(c.ID().AsAddress(), txID)
	c.log.Debugf("ChainNodeConnImplementation::PullTransactionInclusionState(txID=%v)... Done", txIDStr)
}

func (c *chainObj) PullConfirmedOutput(outputID ledgerstate.OutputID) {
	outputIDStr := iscp.OID(outputID)
	c.log.Debugf("ChainNodeConnImplementation::PullConfirmedOutput(outputID=%v)...", outputIDStr)
	c.nodeConn.PullConfirmedOutput(c.ID().AsAddress(), outputID)
	c.log.Debugf("ChainNodeConnImplementation::PullConfirmedOutput(outputID=%v)... Done", outputIDStr)
}

func (c *chainObj) PostTransaction(tx *ledgerstate.Transaction) {
	txIDStr := tx.ID().Base58()
	c.log.Debugf("ChainNodeConnImplementation::PostTransaction(txID=%s)...", txIDStr)
	c.nodeConn.PostTransaction(tx)
	c.log.Debugf("ChainNodeConnImplementation::PostTransaction(txID=%s)... Done", txIDStr)
}
