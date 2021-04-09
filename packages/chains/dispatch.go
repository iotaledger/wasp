package chains

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/txstream"
	"github.com/iotaledger/wasp/packages/coretypes"
)

func (c *Chains) dispatchMsgTransaction(msg *txstream.MsgTransaction) {
	aliasAddr, ok := msg.Address.(*ledgerstate.AliasAddress)
	if !ok {
		c.log.Warnf("chains: cannot dispatch transaction message to non-alias address")
		return
	}
	chainID := coretypes.NewChainID(aliasAddr)
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

func (c *Chains) dispatchMsgInclusionState(msg *txstream.MsgTxInclusionState) {
	aliasAddr, ok := msg.Address.(*ledgerstate.AliasAddress)
	if !ok {
		c.log.Warnf("chains: cannot dispatch inclusion state message to non-alias address")
		return
	}
	chainID := coretypes.NewChainID(aliasAddr)
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
	chainID := coretypes.NewChainID(aliasAddr)
	chain := c.Get(chainID)
	if chain == nil {
		// not interested in this chainID
		return
	}
	c.log.Debugw("dispatch output",
		"outputID", coretypes.OID(msg.Output.ID()),
		"chainid", chainID.String(),
	)
	chain.ReceiveOutput(msg.Output)
}
