package chains

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/txstream"
)

func (c *Chains) dispatchTransactionMsg(msg *txstream.MsgTransaction) {
	c.log.Debugf("NodeConnImplementation::dispatchTransactionMsg...")
	defer c.log.Debugf("NodeConnImplementation::dispatchTransactionMsg... Done")
	aliasAddr, ok := msg.Address.(*iotago.AliasAddress)
	if !ok {
		c.log.Warnf("chains: cannot dispatch transaction message to non-alias address")
		return
	}
	chainID := iscp.ChainIDFromAliasID(aliasAddr)
	chain := c.Get(chainID)
	if chain == nil {
		// not interested in this chainID
		return
	}
	c.log.Debugw("dispatch transaction",
		"txid", msg.Tx.ID().Base58(),
		"chainid", chainID.String(),
	)
	chain.ReceiveTransaction(msg.Tx)
}

func (c *Chains) dispatchInclusionStateMsg(msg *txstream.MsgTxInclusionState) {
	c.log.Debugf("NodeConnImplementation::dispatchInclusionStateMsg...")
	defer c.log.Debugf("NodeConnImplementation::dispatchInclusionStateMsg... Done")
	aliasAddr, ok := msg.Address.(*iotago.AliasAddress)
	if !ok {
		c.log.Warnf("chains: cannot dispatch inclusion state message to non-alias address")
		return
	}
	chainID := iscp.ChainIDFromAliasID(aliasAddr)
	chain := c.Get(chainID)
	if chain == nil {
		// not interested in this chainID
		return
	}
	c.log.Debugw("dispatch transaction",
		"txid", msg.TxID.Base58(),
		"chainid", chainID.String(),
		"inclusion", msg.State.String(),
	)
	chain.ReceiveInclusionState(msg.TxID, msg.State)
}

func (c *Chains) dispatchOutputMsg(msg *txstream.MsgOutput) {
	c.log.Debugf("NodeConnImplementation::dispatchOutputMsg...")
	defer c.log.Debugf("NodeConnImplementation::dispatchOutputMsg... Done")
	aliasAddr, ok := msg.Address.(*iotago.AliasAddress)
	if !ok {
		c.log.Warnf("chains: cannot dispatch output message to non-alias address")
		return
	}
	chainID := iscp.ChainIDFromAliasID(aliasAddr)
	chain := c.Get(chainID)
	if chain == nil {
		// not interested in this message
		return
	}
	c.log.Debugw("dispatch output",
		"outputID", iscp.OID(msg.Output.ID()),
		"chainid", chainID.String(),
	)
	chain.ReceiveOutput(msg.Output)
}

func (c *Chains) dispatchUnspentAliasOutputMsg(msg *txstream.MsgUnspentAliasOutput) {
	c.log.Debugf("NodeConnImplementation::dispatchUnspentAliasOutputMsg...")
	defer c.log.Debugf("NodeConnImplementation::dispatchUnspentAliasOutputMsg... Done")
	chainID := iscp.ChainIDFromAliasID(msg.AliasAddress)
	chain := c.Get(chainID)
	if chain == nil {
		// not interested in this message
		return
	}
	c.log.Debugw("dispatch state",
		"outputID", iscp.OID(msg.AliasOutput.ID()),
		"chainid", chainID.String(),
	)
	chain.ReceiveState(msg.AliasOutput, msg.Timestamp)
}
