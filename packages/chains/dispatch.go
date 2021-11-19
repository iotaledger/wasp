package chains

import (
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/iscp"
)

func (c *Chains) dispatchTransactionMsg(address *ledgerstate.AliasAddress, tx *ledgerstate.Transaction) {
	c.log.Debugf("Chains::dispatchTransactionMsg tx ID %v to address %v", tx.ID().Base58(), address.String())
	chainID := iscp.NewChainID(address)
	chain := c.Get(chainID)
	if chain == nil {
		c.log.Warnf("Chains::dispatchTransactionMsg: not interested in tx ID %v, ignoring it", tx.ID().Base58())
		return
	}
	c.log.Debugf("Chains::dispatchTransactionMsg: dispatching tx ID %v to chain ID %v", tx.ID().Base58(), chainID.String())
	chain.ReceiveTransaction(tx)
	c.log.Debugf("Chains::dispatchTransactionMsg: dispatching tx ID %v completed", tx.ID().Base58())
}

func (c *Chains) dispatchInclusionStateMsg(address *ledgerstate.AliasAddress, txID ledgerstate.TransactionID, iState ledgerstate.InclusionState) {
	c.log.Debugf("Chains::dispatchInclusionStateMsg tx ID %v inclusion state %v to address %v", txID.Base58(), iState.String(), address.String())
	chainID := iscp.NewChainID(address)
	chain := c.Get(chainID)
	if chain == nil {
		c.log.Warnf("Chains::dispatchInclusionStateMsg: not interested in tx ID %v inclusion state, ignoring it", txID.Base58())
		return
	}
	c.log.Debugf("Chains::dispatchInclusionStateMsg: dispatching tx ID %v inclusion state to chain ID %v", txID.Base58(), chainID.String())
	chain.ReceiveInclusionState(txID, iState)
	c.log.Debugf("Chains::dispatchInclusionStateMsg: dispatching tx ID %v inclusion state completed", txID.Base58())
}

func (c *Chains) dispatchOutputMsg(address *ledgerstate.AliasAddress, output ledgerstate.Output) {
	outputID := iscp.OID(output.ID())
	c.log.Debugf("Chains::dispatchOutputMsg output ID %v to address %v", outputID, address.String())
	chainID := iscp.NewChainID(address)
	chain := c.Get(chainID)
	if chain == nil {
		c.log.Warnf("Chains::dispatchOutputMsg: not interested in output ID %v, ignoring it", outputID)
		return
	}
	c.log.Debugf("Chains::dispatchOutputMsg: dispatching output ID %v to chain ID %v", outputID, chainID.String())
	chain.ReceiveOutput(output)
	c.log.Debugf("Chains::dispatchOutputMsg: dispatching output ID %v completed", outputID)
}

func (c *Chains) dispatchUnspentAliasOutputMsg(address *ledgerstate.AliasAddress, output *ledgerstate.AliasOutput, timestamp time.Time) {
	outputID := iscp.OID(output.ID())
	c.log.Debugf("Chains::dispatchUnspentAliasOutputMsg output ID %v timestamp %v to address %v", outputID, timestamp, address.String())
	chainID := iscp.NewChainID(address)
	chain := c.Get(chainID)
	if chain == nil {
		c.log.Warnf("Chains::dispatchUnspentAliasOutputMsg: not interested in output ID %v, ignoring it", outputID)
		return
	}
	c.log.Debugf("Chains::dispatchUnspentAliasOutputMsg: dispatching output ID %v to chain ID %v", outputID, chainID.String())
	chain.ReceiveState(output, timestamp)
	c.log.Debugf("Chains::dispatchOutputMsg: dispatching output ID %v completed", outputID)
}
