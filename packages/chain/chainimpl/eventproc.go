// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainimpl

import (
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/peering"
	//	"github.com/iotaledger/wasp/packages/hashing"
	//	"github.com/iotaledger/wasp/packages/iscp"
	//	"github.com/iotaledger/wasp/packages/state"
	//	"github.com/iotaledger/wasp/packages/util"
)

func (c *chainObj) handleMessagesLoop() {
	nDismissChainMsgChannel := c.dismissChainMsgChannel.Out()
	nStateMsgChannel := c.stateMsgChannel.Out()
	nOffLedgerRequestPeerMsgChannel := c.offLedgerRequestPeerMsgChannel.Out()
	nRequestAckPeerMsgChannel := c.requestAckPeerMsgChannel.Out()
	nMissingRequestIDsPeerMsgChannel := c.missingRequestIDsPeerMsgChannel.Out()
	nMissingRequestPeerMsgChannel := c.missingRequestPeerMsgChannel.Out()
	nTimerTickMsgChannel := c.timerTickMsgChannel.Out()
	for {
		select {
		case msg, ok := <-nDismissChainMsgChannel:
			if ok {
				c.handleDismissChain(msg.(messages.DismissChainMsg))
			} else {
				nDismissChainMsgChannel = nil
			}
		case msg, ok := <-nStateMsgChannel:
			if ok {
				c.handleLedgerState(msg.(*messages.StateMsg))
			} else {
				nStateMsgChannel = nil
			}
		case msg, ok := <-nOffLedgerRequestPeerMsgChannel:
			if ok {
				c.handleOffLedgerRequestPeerMsg(msg.(*peering.PeerMessage))
			} else {
				nOffLedgerRequestPeerMsgChannel = nil
			}
		case msg, ok := <-nRequestAckPeerMsgChannel:
			if ok {
				c.handleRequestAckPeerMsg(msg.(*peering.PeerMessage))
			} else {
				nRequestAckPeerMsgChannel = nil
			}
		case msg, ok := <-nMissingRequestIDsPeerMsgChannel:
			if ok {
				c.handleMissingRequestIDsPeerMsg(msg.(*peering.PeerMessage))
			} else {
				nMissingRequestIDsPeerMsgChannel = nil
			}
		case msg, ok := <-nMissingRequestPeerMsgChannel:
			if ok {
				c.handleMissingRequestPeerMsg(msg.(*peering.PeerMessage))
			} else {
				nMissingRequestPeerMsgChannel = nil
			}
		case msg, ok := <-nTimerTickMsgChannel:
			if ok {
				c.handleTimerTick(msg.(messages.TimerTick))
			} else {
				nTimerTickMsgChannel = nil
			}
		}
		if nDismissChainMsgChannel == nil &&
			nStateMsgChannel == nil &&
			nOffLedgerRequestPeerMsgChannel == nil &&
			nRequestAckPeerMsgChannel == nil &&
			nMissingRequestIDsPeerMsgChannel == nil &&
			nMissingRequestPeerMsgChannel == nil &&
			nTimerTickMsgChannel == nil {
			return
		}
	}
}

func (c *chainObj) EnqueueDismissChain(reason string) {
	c.dismissChainMsgChannel.In() <- messages.DismissChainMsg{Reason: reason}
}

func (c *chainObj) handleDismissChain(msg messages.DismissChainMsg) {
	c.Dismiss(msg.Reason)
}

func (c *chainObj) enqueueMissingRequestIDsPeerMsg(peerMsg *peering.PeerMessage) {
	c.missingRequestIDsPeerMsgChannel.In() <- peerMsg
}

func (c *chainObj) handleMissingRequestIDsPeerMsg(peerMsg *peering.PeerMessage) {
	if !c.pullMissingRequestsFromCommittee {
		return
	}
	msg, err := messages.MissingRequestIDsMsgFromBytes(peerMsg.MsgData)
	if err != nil {
		c.log.Error(err)
		return
	}
	c.SendMissingRequestsToPeer(msg, peerMsg.SenderNetID)
}

func (c *chainObj) enqueueOffLedgerRequestPeerMsg(peerMsg *peering.PeerMessage) {
	c.offLedgerRequestPeerMsgChannel.In() <- peerMsg
}

func (c *chainObj) handleOffLedgerRequestPeerMsg(peerMsg *peering.PeerMessage) {
	msg, err := messages.OffLedgerRequestMsgFromBytes(peerMsg.MsgData)
	if err != nil {
		c.log.Error(err)
		return
	}
	c.ReceiveOffLedgerRequest(msg.Req, peerMsg.SenderNetID)
}

func (c *chainObj) enqueueRequestAckPeerMsg(peerMsg *peering.PeerMessage) {
	c.requestAckPeerMsgChannel.In() <- peerMsg
}

func (c *chainObj) handleRequestAckPeerMsg(peerMsg *peering.PeerMessage) {
	msg, err := messages.RequestAckMsgFromBytes(peerMsg.MsgData)
	if err != nil {
		c.log.Error(err)
		return
	}
	c.ReceiveRequestAckMessage(msg.ReqID, peerMsg.SenderNetID)
}

func (c *chainObj) enqueueMissingRequestPeerMsg(peerMsg *peering.PeerMessage) {
	c.missingRequestPeerMsgChannel.In() <- peerMsg
}

func (c *chainObj) handleMissingRequestPeerMsg(peerMsg *peering.PeerMessage) {
	if !c.pullMissingRequestsFromCommittee {
		return
	}
	msg, err := messages.MissingRequestMsgFromBytes(peerMsg.MsgData)
	if err != nil {
		c.log.Error(err)
		return
	}
	if c.consensus.ShouldReceiveMissingRequest(msg.Request) {
		c.mempool.ReceiveRequest(msg.Request)
	}
}

func (c *chainObj) enqueueLedgerState(chainOutput *ledgerstate.AliasOutput, timestamp time.Time) {
	c.stateMsgChannel.In() <- &messages.StateMsg{
		ChainOutput: chainOutput,
		Timestamp:   timestamp,
	}
}

func (c *chainObj) handleLedgerState(msg *messages.StateMsg) {
	c.processStateMessage(msg)
}

func (c *chainObj) enqueueTimerTick(tick int) {
	c.timerTickMsgChannel.In() <- messages.TimerTick(tick)
}

func (c *chainObj) handleTimerTick(msg messages.TimerTick) {
	if msg%2 == 0 {
		c.stateMgr.EventTimerMsg(msg / 2)
	} else if c.consensus != nil {
		c.consensus.EventTimerMsg(msg / 2)
	}
	if msg%40 == 0 {
		stats := c.mempool.Info()
		c.log.Debugf("mempool total = %d, ready = %d, in = %d, out = %d", stats.TotalPool, stats.ReadyCounter, stats.InPoolCounter, stats.OutPoolCounter)
	}
}

/*// EventGetBlockMsg is a request for a block while syncing
func (sm *stateManager) EventGetBlockMsg(msg *messages.GetBlockMsg) {
	sm.eventGetBlockMsgCh <- msg
}

func (sm *stateManager) eventGetBlockMsg(msg *messages.GetBlockMsg) {
	sm.log.Debugw("EventGetBlockMsg received: ",
		"sender", msg.SenderNetID,
		"block index", msg.BlockIndex,
	)
	if sm.stateOutput == nil { // Not a necessary check, only for optimization.
		sm.log.Debugf("EventGetBlockMsg ignored: stateOutput is nil")
		return
	}
	if msg.BlockIndex > sm.stateOutput.GetStateIndex() { // Not a necessary check, only for optimization.
		sm.log.Debugf("EventGetBlockMsg ignored 1: block #%d not found. Current state index: #%d",
			msg.BlockIndex, sm.stateOutput.GetStateIndex())
		return
	}
	blockBytes, err := state.LoadBlockBytes(sm.store, msg.BlockIndex)
	if err != nil {
		sm.log.Errorf("EventGetBlockMsg: LoadBlockBytes: %v", err)
		return
	}
	if blockBytes == nil {
		sm.log.Debugf("EventGetBlockMsg ignored 2: block #%d not found. Current state index: #%d",
			msg.BlockIndex, sm.stateOutput.GetStateIndex())
		return
	}

	sm.log.Debugf("EventGetBlockMsg for state index #%d --> responding to peer %s", msg.BlockIndex, msg.SenderNetID)

	sm.peers.SendSimple(msg.SenderNetID, messages.MsgBlock, util.MustBytes(&messages.BlockMsg{
		BlockBytes: blockBytes,
	}))
}

// EventBlockMsg
func (sm *stateManager) EventBlockMsg(msg *messages.BlockMsg) {
	sm.eventBlockMsgCh <- msg
}

func (sm *stateManager) eventBlockMsg(msg *messages.BlockMsg) {
	sm.log.Debugf("EventBlockMsg received from %v", msg.SenderNetID)
	if sm.stateOutput == nil {
		sm.log.Debugf("EventBlockMsg ignored: stateOutput is nil")
		return
	}
	block, err := state.BlockFromBytes(msg.BlockBytes)
	if err != nil {
		sm.log.Warnf("EventBlockMsg ignored: wrong block received from peer %s. Err: %v", msg.SenderNetID, err)
		return
	}
	sm.log.Debugw("EventBlockMsg from ",
		"sender", msg.SenderNetID,
		"block index", block.BlockIndex(),
		"approving output", iscp.OID(block.ApprovingOutputID()),
	)
	if sm.addBlockFromPeer(block) {
		sm.takeAction()
	}
}

func (sm *stateManager) EventOutputMsg(msg ledgerstate.Output) {
	sm.eventOutputMsgCh <- msg
}

func (sm *stateManager) eventOutputMsg(msg ledgerstate.Output) {
	sm.log.Debugf("EventOutputMsg received: %s", iscp.OID(msg.ID()))
	chainOutput, ok := msg.(*ledgerstate.AliasOutput)
	if !ok {
		sm.log.Debugf("EventOutputMsg ignored: output is of type %t, expecting *ledgerstate.AliasOutput", msg)
		return
	}
	if sm.outputPulled(chainOutput) {
		sm.takeAction()
	}
}

// EventStateTransactionMsg triggered whenever new state transaction arrives
// the state transaction may be confirmed or not
func (sm *stateManager) EventStateMsg(msg *messages.StateMsg) {
	sm.eventStateOutputMsgCh <- msg
}

func (sm *stateManager) eventStateMsg(msg *messages.StateMsg) {
	sm.log.Debugw("EventStateMsg received: ",
		"state index", msg.ChainOutput.GetStateIndex(),
		"chainOutput", iscp.OID(msg.ChainOutput.ID()),
	)
	stateHash, err := hashing.HashValueFromBytes(msg.ChainOutput.GetStateData())
	if err != nil {
		sm.log.Errorf("EventStateMsg ignored: failed to parse state hash: %v", err)
		return
	}
	sm.log.Debugf("EventStateMsg state hash is %v", stateHash.String())
	if sm.stateOutputReceived(msg.ChainOutput, msg.Timestamp) {
		sm.takeAction()
	}
}

func (sm *stateManager) EventStateCandidateMsg(state state.VirtualStateAccess, outputID ledgerstate.OutputID) {
	sm.eventStateCandidateMsgCh <- &messages.StateCandidateMsg{
		State:             state,
		ApprovingOutputID: outputID,
	}
}

func (sm *stateManager) eventStateCandidateMsg(msg *messages.StateCandidateMsg) {
	sm.log.Debugf("EventStateCandidateMsg received: state index: %d, timestamp: %v",
		msg.State.BlockIndex(), msg.State.Timestamp(),
	)
	if sm.stateOutput == nil {
		sm.log.Debugf("EventStateCandidateMsg ignored: stateOutput is nil")
		return
	}
	if sm.addStateCandidateFromConsensus(msg.State, msg.ApprovingOutputID) {
		sm.takeAction()
	}
}

func (sm *stateManager) EventTimerMsg(msg messages.TimerTick) {
	if msg%2 == 0 {
		sm.eventTimerMsgCh <- msg
	}
}

func (sm *stateManager) eventTimerMsg() {
	sm.log.Debugf("EventTimerMsg received")
	sm.takeAction()
}
*/
