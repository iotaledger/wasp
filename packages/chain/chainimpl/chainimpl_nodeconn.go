// Provides implementations for chain.ChainNodeConnection methods
package chainimpl

import (
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/iscp"
)

func (c *chainObj) AttachToTransactionReceived(fun chain.NodeConnectionHandleTransactionFun) {
	c.nodeConn.AttachToTransactionReceived(c.ID().AsAliasAddress(), func(tx *ledgerstate.Transaction) {
		chain.CountMessageStats(&c.stats.InTransaction, tx)
		fun(tx)
	})
}

func (c *chainObj) AttachToInclusionStateReceived(fun chain.NodeConnectionHandleInclusionStateFun) {
	c.nodeConn.AttachToInclusionStateReceived(c.ID().AsAliasAddress(), func(txID ledgerstate.TransactionID, iState ledgerstate.InclusionState) {
		chain.CountMessageStats(&c.stats.InInclusionState, struct {
			TransactionID  ledgerstate.TransactionID
			InclusionState ledgerstate.InclusionState
		}{
			TransactionID:  txID,
			InclusionState: iState,
		})
		fun(txID, iState)
	})
}

func (c *chainObj) AttachToOutputReceived(fun chain.NodeConnectionHandleOutputFun) {
	c.nodeConn.AttachToOutputReceived(c.ID().AsAliasAddress(), func(output ledgerstate.Output) {
		chain.CountMessageStats(&c.stats.InOutput, output)
		fun(output)
	})
}

func (c *chainObj) AttachToUnspentAliasOutputReceived(fun chain.NodeConnectionHandleUnspentAliasOutputFun) {
	c.nodeConn.AttachToUnspentAliasOutputReceived(c.ID().AsAliasAddress(), func(output *ledgerstate.AliasOutput, timestamp time.Time) {
		chain.CountMessageStats(&c.stats.InUnspentAliasOutput, struct {
			AliasOutput *ledgerstate.AliasOutput
			Timestamp   time.Time
		}{
			AliasOutput: output,
			Timestamp:   timestamp,
		})
		fun(output, timestamp)
	})
}

func (c *chainObj) PullState() {
	c.log.Debugf("ChainNodeConnection::PullState...")
	chain.CountMessageStats(&c.stats.OutPullState, nil)
	c.nodeConn.PullState(c.ID().AsAliasAddress())
	c.log.Debugf("ChainNodeConnection::PullState... Done")
}

func (c *chainObj) PullTransactionInclusionState(txID ledgerstate.TransactionID) {
	txIDStr := txID.Base58()
	c.log.Debugf("ChainNodeConnImplementation::PullTransactionInclusionState(txID=%v)...", txIDStr)
	chain.CountMessageStats(&c.stats.OutPullTransactionInclusionState, txID)
	c.nodeConn.PullTransactionInclusionState(c.ID().AsAddress(), txID)
	c.log.Debugf("ChainNodeConnImplementation::PullTransactionInclusionState(txID=%v)... Done", txIDStr)
}

func (c *chainObj) PullConfirmedOutput(outputID ledgerstate.OutputID) {
	outputIDStr := iscp.OID(outputID)
	c.log.Debugf("ChainNodeConnImplementation::PullConfirmedOutput(outputID=%v)...", outputIDStr)
	chain.CountMessageStats(&c.stats.OutPullConfirmedOutput, outputID)
	c.nodeConn.PullConfirmedOutput(c.ID().AsAddress(), outputID)
	c.log.Debugf("ChainNodeConnImplementation::PullConfirmedOutput(outputID=%v)... Done", outputIDStr)
}

func (c *chainObj) PostTransaction(tx *ledgerstate.Transaction) {
	txIDStr := tx.ID().Base58()
	c.log.Debugf("ChainNodeConnImplementation::PostTransaction(txID=%s)...", txIDStr)
	chain.CountMessageStats(&c.stats.OutPostTransaction, tx)
	c.nodeConn.PostTransaction(tx)
	c.log.Debugf("ChainNodeConnImplementation::PostTransaction(txID=%s)... Done", txIDStr)
}
