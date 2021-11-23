// Provides implementations for chain.ChainNodeConnection methods
package nodeconnimpl

import (
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/iscp"
)

type chainNodeConnImplementation struct {
	chainID  *iscp.ChainID
	nodeConn chain.NodeConnection
	stats    *chain.NodeConnectionMessagesStats
	log      *logger.Logger // logger for one chain
}

var _ chain.ChainNodeConnection = &chainNodeConnImplementation{}

func NewChainNodeConnection(chainID *iscp.ChainID, nodeConn chain.NodeConnection, log *logger.Logger) chain.ChainNodeConnection {
	return &chainNodeConnImplementation{
		chainID:  chainID,
		nodeConn: nodeConn,
		stats:    &chain.NodeConnectionMessagesStats{},
		log:      log,
	}
}

func (c *chainNodeConnImplementation) AttachToTransactionReceived(fun chain.NodeConnectionHandleTransactionFun) {
	c.nodeConn.AttachToTransactionReceived(c.chainID.AsAliasAddress(), func(tx *ledgerstate.Transaction) {
		chain.CountMessageStats(&c.stats.InTransaction, tx)
		fun(tx)
	})
}

func (c *chainNodeConnImplementation) AttachToInclusionStateReceived(fun chain.NodeConnectionHandleInclusionStateFun) {
	c.nodeConn.AttachToInclusionStateReceived(c.chainID.AsAliasAddress(), func(txID ledgerstate.TransactionID, iState ledgerstate.InclusionState) {
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

func (c *chainNodeConnImplementation) AttachToOutputReceived(fun chain.NodeConnectionHandleOutputFun) {
	c.nodeConn.AttachToOutputReceived(c.chainID.AsAliasAddress(), func(output ledgerstate.Output) {
		chain.CountMessageStats(&c.stats.InOutput, output)
		fun(output)
	})
}

func (c *chainNodeConnImplementation) AttachToUnspentAliasOutputReceived(fun chain.NodeConnectionHandleUnspentAliasOutputFun) {
	c.nodeConn.AttachToUnspentAliasOutputReceived(c.chainID.AsAliasAddress(), func(output *ledgerstate.AliasOutput, timestamp time.Time) {
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

func (c *chainNodeConnImplementation) PullState() {
	c.log.Debugf("ChainNodeConnection::PullState...")
	chain.CountMessageStats(&c.stats.OutPullState, nil)
	c.nodeConn.PullState(c.chainID.AsAliasAddress())
	c.log.Debugf("ChainNodeConnection::PullState... Done")
}

func (c *chainNodeConnImplementation) PullTransactionInclusionState(txID ledgerstate.TransactionID) {
	txIDStr := txID.Base58()
	c.log.Debugf("ChainNodeConnImplementation::PullTransactionInclusionState(txID=%v)...", txIDStr)
	chain.CountMessageStats(&c.stats.OutPullTransactionInclusionState, txID)
	c.nodeConn.PullTransactionInclusionState(c.chainID.AsAddress(), txID)
	c.log.Debugf("ChainNodeConnImplementation::PullTransactionInclusionState(txID=%v)... Done", txIDStr)
}

func (c *chainNodeConnImplementation) PullConfirmedOutput(outputID ledgerstate.OutputID) {
	outputIDStr := iscp.OID(outputID)
	c.log.Debugf("ChainNodeConnImplementation::PullConfirmedOutput(outputID=%v)...", outputIDStr)
	chain.CountMessageStats(&c.stats.OutPullConfirmedOutput, outputID)
	c.nodeConn.PullConfirmedOutput(c.chainID.AsAddress(), outputID)
	c.log.Debugf("ChainNodeConnImplementation::PullConfirmedOutput(outputID=%v)... Done", outputIDStr)
}

func (c *chainNodeConnImplementation) PostTransaction(tx *ledgerstate.Transaction) {
	txIDStr := tx.ID().Base58()
	c.log.Debugf("ChainNodeConnImplementation::PostTransaction(txID=%s)...", txIDStr)
	chain.CountMessageStats(&c.stats.OutPostTransaction, tx)
	c.nodeConn.PostTransaction(tx)
	c.log.Debugf("ChainNodeConnImplementation::PostTransaction(txID=%s)... Done", txIDStr)
}

func (c *chainNodeConnImplementation) GetStats() chain.NodeConnectionMessagesStats {
	return *c.stats
}
