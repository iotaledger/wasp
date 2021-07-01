package chains

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/txstream"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/chainid"
)

func (c *Chains) dispatchTransactionMsg(msg *txstream.MsgTransaction) {
	aliasAddr, ok := msg.Address.(*ledgerstate.AliasAddress)
	if !ok {
		c.log.Warnf("chains: cannot dispatch transaction message to non-alias address")
		return
	}
	chainID := chainid.NewChainID(aliasAddr)
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
	aliasAddr, ok := msg.Address.(*ledgerstate.AliasAddress)
	if !ok {
		c.log.Warnf("chains: cannot dispatch inclusion state message to non-alias address")
		return
	}
	chainID := chainid.NewChainID(aliasAddr)
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
	aliasAddr, ok := msg.Address.(*ledgerstate.AliasAddress)
	if !ok {
		c.log.Warnf("chains: cannot dispatch output message to non-alias address")
		return
	}
	chainID := chainid.NewChainID(aliasAddr)
	chain := c.Get(chainID)
	if chain == nil {
		// not interested in this message
		return
	}
	c.log.Debugw("dispatch output",
		"outputID", coretypes.OID(msg.Output.ID()),
		"chainid", chainID.String(),
	)
	chain.ReceiveOutput(msg.Output)
}

func (c *Chains) dispatchUnspentAliasOutputMsg(msg *txstream.MsgUnspentAliasOutput) {
	chainID := chainid.NewChainID(msg.AliasAddress)
	chain := c.Get(chainID)
	if chain == nil {
		// not interested in this message
		return
	}
	c.log.Debugw("dispatch state",
		"outputID", coretypes.OID(msg.AliasOutput.ID()),
		"chainid", chainID.String(),
	)
	chain.ReceiveState(msg.AliasOutput, msg.Timestamp)
}
