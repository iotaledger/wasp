// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Provides implementations for chain.ChainEntry methods
package chainimpl

import (
	"github.com/iotaledger/wasp/packages/publisher"
)

func (c *chainObj) Dismiss(reason string) {
	c.log.Infof("Dismiss chain. Reason: '%s'", reason)

	c.dismissOnce.Do(func() {
		c.dismissed.Store(true)
		c.chainPeers.Detach(c.receiveChainPeerMessagesAttachID)
		c.nodeConn.DetachFromOnLedgerRequest()
		c.nodeConn.DetachFromAliasOutput()
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
		c.aliasOutputPipe.Close()
		c.offLedgerRequestPeerMsgPipe.Close()
		c.missingRequestIDsPeerMsgPipe.Close()
		c.missingRequestPeerMsgPipe.Close()
		c.timerTickMsgPipe.Close()
	})

	publisher.Publish("dismissed_chain", c.chainID.String())
	c.log.Debug("Chain dismissed")
}

func (c *chainObj) IsDismissed() bool {
	return c.dismissed.Load()
}
