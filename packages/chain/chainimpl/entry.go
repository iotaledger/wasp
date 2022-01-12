// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Provides implementations for chain.ChainEntry methods
package chainimpl

import (
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/packages/transaction"
)

func (c *chainObj) ReceiveTransaction(tx *ledgerstate.Transaction) {
	c.log.Debugf("ReceiveTransaction: %s", tx.ID().Base58())
	reqs, err := request.OnLedgerFromTransaction(tx, c.chainID.AsAddress())
	if err != nil {
		c.log.Warnf("failed to parse transaction %s: %v", tx.ID().Base58(), err)
		return
	}
	for _, req := range reqs {
		c.mempool.ReceiveRequest(req)
	}
	if chainOut := transaction.GetAliasOutput(tx, c.chainID.AsAddress()); chainOut != nil {
		c.ReceiveState(chainOut, tx.Essence().Timestamp())
	}
}

func (c *chainObj) ReceiveState(stateOutput *ledgerstate.AliasOutput, timestamp time.Time) {
	c.log.Debugf("ReceiveState #%d: outputID: %s, stateAddr: %s",
		stateOutput.GetStateIndex(), iscp.OID(stateOutput.ID()), stateOutput.GetStateAddress().Base58())
	c.EnqueueLedgerState(stateOutput, timestamp)
}

func (c *chainObj) Dismiss(reason string) {
	c.log.Infof("Dismiss chain. Reason: '%s'", reason)

	c.dismissOnce.Do(func() {
		c.dismissed.Store(true)
		c.chainPeers.Detach(c.receiveChainPeerMessagesAttachID)
		c.nodeConn.DetachFromUnspentAliasOutputReceived()
		c.nodeConn.DetachFromTransactionReceived()
		c.eventChainTransition.Detach(c.eventChainTransitionClosure)

		c.mempool.Close()
		c.stateMgr.Close()
		cmt := c.getCommittee()
		if cmt != nil {
			c.detachFromCommitteePeerMessagesFun()
			cmt.Close()
		}
		if c.consensus != nil {
			c.consensus.Close()
		}

		c.eventRequestProcessed.DetachAll()
		c.eventChainTransition.DetachAll()
		c.chainPeers.Close()
		c.nodeConn.Close()

		c.dismissChainMsgPipe.Close()
		c.stateMsgPipe.Close()
		c.offLedgerRequestPeerMsgPipe.Close()
		c.requestAckPeerMsgPipe.Close()
		c.missingRequestIDsPeerMsgPipe.Close()
		c.missingRequestPeerMsgPipe.Close()
		c.timerTickMsgPipe.Close()
	})

	publisher.Publish("dismissed_chain", c.chainID.Base58())
	c.log.Debug("Chain dismissed")
}

func (c *chainObj) IsDismissed() bool {
	return c.dismissed.Load()
}
