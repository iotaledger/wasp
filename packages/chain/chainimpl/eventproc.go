// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainimpl

import (
	"errors"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/committee"
	"github.com/iotaledger/wasp/packages/chain/consensus"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/request"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"golang.org/x/xerrors"
)

func (c *chainObj) recvLoop() {
	dismissChainMsgChannel := c.dismissChainMsgPipe.Out()
	stateMsgChannel := c.stateMsgPipe.Out()
	offLedgerRequestMsgChannel := c.offLedgerRequestPeerMsgPipe.Out()
	requestAckMsgChannel := c.requestAckPeerMsgPipe.Out()
	missingRequestIDsMsgChannel := c.missingRequestIDsPeerMsgPipe.Out()
	missingRequestMsgChannel := c.missingRequestPeerMsgPipe.Out()
	timerTickMsgChannel := c.timerTickMsgPipe.Out()
	for {
		select {
		case msg, ok := <-dismissChainMsgChannel:
			if ok {
				c.handleDismissChain(msg.(DismissChainMsg))
			} else {
				dismissChainMsgChannel = nil
			}
		case msg, ok := <-stateMsgChannel:
			if ok {
				c.handleLedgerState(msg.(*messages.StateMsg))
			} else {
				stateMsgChannel = nil
			}
		case msg, ok := <-offLedgerRequestMsgChannel:
			if ok {
				c.handleOffLedgerRequestMsg(msg.(*messages.OffLedgerRequestMsgIn))
			} else {
				offLedgerRequestMsgChannel = nil
			}
		case msg, ok := <-requestAckMsgChannel:
			if ok {
				c.handleRequestAckPeerMsg(msg.(*messages.RequestAckMsgIn))
			} else {
				requestAckMsgChannel = nil
			}
		case msg, ok := <-missingRequestIDsMsgChannel:
			if ok {
				c.handleMissingRequestIDsMsg(msg.(*messages.MissingRequestIDsMsgIn))
			} else {
				missingRequestIDsMsgChannel = nil
			}
		case msg, ok := <-missingRequestMsgChannel:
			if ok {
				c.handleMissingRequestMsg(msg.(*messages.MissingRequestMsg))
			} else {
				missingRequestMsgChannel = nil
			}
		case msg, ok := <-timerTickMsgChannel:
			if ok {
				c.handleTimerTick(msg.(messages.TimerTick))
			} else {
				timerTickMsgChannel = nil
			}
		}
		if dismissChainMsgChannel == nil &&
			stateMsgChannel == nil &&
			offLedgerRequestMsgChannel == nil &&
			requestAckMsgChannel == nil &&
			missingRequestIDsMsgChannel == nil &&
			missingRequestMsgChannel == nil &&
			timerTickMsgChannel == nil {
			return
		}
	}
}

func (c *chainObj) EnqueueDismissChain(reason string) {
	c.dismissChainMsgPipe.In() <- DismissChainMsg{Reason: reason}
	c.chainMetrics.CountMessages()
}

func (c *chainObj) handleDismissChain(msg DismissChainMsg) {
	c.log.Debugf("handleDismissChain message received, reason=%s", msg.Reason)
	c.Dismiss(msg.Reason)
}

func (c *chainObj) EnqueueLedgerState(chainOutput *iotago.AliasOutput, timestamp time.Time) {
	c.stateMsgPipe.In() <- &messages.StateMsg{
		ChainOutput: chainOutput,
		Timestamp:   timestamp,
	}
	c.chainMetrics.CountMessages()
}

// handleLedgerState processes the only chain output which exists on the chain's address
// If necessary, it creates/changes/rotates committee object
func (c *chainObj) handleLedgerState(msg *messages.StateMsg) {
	c.log.Debugf("handleLedgerState message received, stateIndex: %d, stateAddr: %s, state transition: %v, timestamp: %v",
		msg.ChainOutput.GetStateIndex(), msg.ChainOutput.GetStateAddress().Base58(), !msg.ChainOutput.GetIsGovernanceUpdated(), msg.Timestamp)
	sh, err := hashing.HashValueFromBytes(msg.ChainOutput.GetStateData())
	if err != nil {
		c.log.Error(xerrors.Errorf("parsing state hash: %w", err))
		return
	}
	c.log.Debugf("handleLedgerState stateHash: %s", sh.String())
	cmt := c.getCommittee()

	if cmt != nil {
		err = c.rotateCommitteeIfNeeded(msg.ChainOutput, cmt)
	} else {
		err = c.createCommitteeIfNeeded(msg.ChainOutput)
	}
	if err != nil {
		c.log.Errorf("processStateMessage: %v", err)
		return
	}
	c.stateMgr.EnqueueStateMsg(msg)
	c.log.Debugf("handleLedgerState passed to state manager")
}

func (c *chainObj) rotateCommitteeIfNeeded(anchorOutput *ledgerstate.AliasOutput, currentCmt chain.Committee) error {
	if currentCmt.Address().Equals(anchorOutput.GetStateAddress()) {
		// nothing changed. no rotation
		return nil
	}
	// address changed
	if !anchorOutput.GetIsGovernanceUpdated() {
		return xerrors.Errorf("rotateCommitteeIfNeeded: inconsistency. Governance transition expected... New output: %s", anchorOutput.String())
	}
	dkShare, err := c.getChainDKShare(anchorOutput.GetStateAddress())
	if err != nil {
		if !errors.Is(err, registry.ErrDKShareNotFound) {
			return xerrors.Errorf("rotateCommitteeIfNeeded: unable to load dkShare: %w", err)
		}
	}
	// rotation needed
	// close current in any case
	c.log.Infof("CLOSING COMMITTEE for %s", currentCmt.Address().Base58())

	currentCmt.Close()
	c.consensus.Close()
	c.setCommittee(nil)
	c.consensus = nil
	if dkShare != nil {
		// create new if committee record is available
		if err = c.createNewCommitteeAndConsensus(dkShare); err != nil {
			return xerrors.Errorf("rotateCommitteeIfNeeded: creating committee and consensus: %v", err)
		}
	}
	return nil
}

func (c *chainObj) createCommitteeIfNeeded(anchorOutput *ledgerstate.AliasOutput) error {
	// check if I am in the committee
	dkShare, err := c.getChainDKShare(anchorOutput.GetStateAddress())
	if err != nil {
		if errors.Is(err, registry.ErrDKShareNotFound) {
			return nil
		}
		return xerrors.Errorf("createCommitteeIfNeeded: unable to load dkShare: %w", err)
	}
	if dkShare != nil {
		// create if record is present
		if err = c.createNewCommitteeAndConsensus(dkShare); err != nil {
			return xerrors.Errorf("createCommitteeIfNeeded: creating committee and consensus: %w", err)
		}
	}
	return nil
}

func (c *chainObj) getChainDKShare(addr ledgerstate.Address) (*tcrypto.DKShare, error) {
	//
	// just in case check if I am among committee nodes
	// should not happen
	selfPubKey := c.netProvider.Self().PubKey()
	cmtDKShare, err := c.dksProvider.LoadDKShare(addr)
	if err != nil {
		return nil, err
	}
	for i := range cmtDKShare.NodePubKeys {
		if *cmtDKShare.NodePubKeys[i] == *selfPubKey {
			return cmtDKShare, nil
		}
	}
	return nil, xerrors.Errorf("createCommitteeIfNeeded: I am not among nodes of the committee record. Inconsistency")
}

func (c *chainObj) createNewCommitteeAndConsensus(dkShare *tcrypto.DKShare) error {
	c.log.Debugf("createNewCommitteeAndConsensus: creating a new committee...")
	if c.detachFromCommitteePeerMessagesFun != nil {
		c.detachFromCommitteePeerMessagesFun()
	}
	cmt, cmtPeerGroup, err := committee.New(
		dkShare,
		c.chainID,
		c.netProvider,
		c.log,
	)
	if err != nil {
		c.setCommittee(nil)
		return xerrors.Errorf("createNewCommitteeAndConsensus: failed to create committee object for state address %s: %w",
			dkShare.Address.Base58(), err)
	}
	attachID := cmtPeerGroup.Attach(peering.PeerMessageReceiverChain, c.receiveCommitteePeerMessages)
	c.detachFromCommitteePeerMessagesFun = func() {
		cmtPeerGroup.Detach(attachID)
	}
	c.log.Debugf("creating new consensus object...")
	c.consensus = consensus.New(c, c.mempool, cmt, cmtPeerGroup, c.nodeConn, c.pullMissingRequestsFromCommittee, c.chainMetrics)
	c.setCommittee(cmt)

	c.log.Infof("NEW COMMITTEE OF VALIDATORS has been initialized for the state address %s", dkShare.Address.Base58())
	return nil
}

func (c *chainObj) EnqueueOffLedgerRequestMsg(msg *messages.OffLedgerRequestMsgIn) {
	c.offLedgerRequestPeerMsgPipe.In() <- msg
	c.chainMetrics.CountMessages()
}

func (c *chainObj) handleOffLedgerRequestMsg(msg *messages.OffLedgerRequestMsgIn) {
	c.log.Debugf("handleOffLedgerRequestMsg message received from peer %v, reqID: %s", msg.SenderPubKey.String(), msg.Req.ID().Base58())
	c.sendRequestAcknowledgementMsg(msg.Req.ID(), msg.SenderPubKey)

	if !c.isRequestValid(msg.Req) {
		// this means some node broadcasted an invalid request (bad chainID or signature)
		// TODO should the sender node be punished somehow?
		c.log.Errorf("handleOffLedgerRequestMsg message ignored: request is not valid")
		return
	}
	if !c.mempool.ReceiveRequest(msg.Req) {
		c.log.Errorf("handleOffLedgerRequestMsg message ignored: mempool hasn't accepted it")
		return
	}
	c.broadcastOffLedgerRequest(msg.Req)
	c.log.Debugf("handleOffLedgerRequestMsg message added to mempool and broadcasted: reqID: %s", msg.Req.ID().Base58())
}

func (c *chainObj) sendRequestAcknowledgementMsg(reqID iscp.RequestID, peerPubKey *ed25519.PublicKey) {
	if peerPubKey == nil {
		return
	}
	c.log.Debugf("sendRequestAcknowledgementMsg: reqID: %s, peerID: %s", reqID.Base58(), peerPubKey.String())
	msg := &messages.RequestAckMsg{ReqID: &reqID}
	c.chainPeers.SendMsgByPubKey(peerPubKey, peering.PeerMessageReceiverChain, chain.PeerMsgTypeRequestAck, msg.Bytes())
}

func (c *chainObj) isRequestValid(req *iscp.OffLedgerRequestData) bool {
	return req.ChainID().Equals(c.ID()) && req.VerifySignature()
}

func (c *chainObj) broadcastOffLedgerRequest(req *request.OffLedger) {
	c.log.Debugf("broadcastOffLedgerRequest: toNPeers: %d, reqID: %s", c.offledgerBroadcastUpToNPeers, req.ID().Base58())
	msg := &messages.OffLedgerRequestMsg{
		ChainID: c.chainID,
		Req:     req,
	}
	cmt := c.getCommittee()
	getPeerPubKeys := c.chainPeers.GetRandomOtherPeers

	if cmt != nil {
		getPeerPubKeys = cmt.GetRandomValidators
	}

	sendMessage := func(ackPeers []*ed25519.PublicKey) {
		peerPubKeys := getPeerPubKeys(c.offledgerBroadcastUpToNPeers)
		for _, peerPubKey := range peerPubKeys {
			if shouldSendToPeer(peerPubKey, ackPeers) {
				c.log.Debugf("sending offledger request ID: reqID: %s, peerPubKey: %s", req.ID().Base58(), peerPubKey.String())
				c.chainPeers.SendMsgByPubKey(peerPubKey, peering.PeerMessageReceiverChain, chain.PeerMsgTypeOffLedgerRequest, msg.Bytes())
			}
		}
	}

	ticker := time.NewTicker(c.offledgerBroadcastInterval)
	stopBroadcast := func() {
		c.offLedgerReqsAcksMutex.Lock()
		delete(c.offLedgerReqsAcks, req.ID())
		c.offLedgerReqsAcksMutex.Unlock()
		ticker.Stop()
	}

	go func() {
		defer stopBroadcast()
		for {
			<-ticker.C
			// check if processed (request already left the mempool)
			if !c.mempool.HasRequest(req.ID()) {
				return
			}
			c.offLedgerReqsAcksMutex.RLock()
			ackPeers := c.offLedgerReqsAcks[(*req).ID()]
			c.offLedgerReqsAcksMutex.RUnlock()
			if cmt != nil && len(ackPeers) >= int(cmt.Size())-1 {
				// this node is part of the committee and the message has already been received by every other committee node
				return
			}
			sendMessage(ackPeers)
		}
	}()
}

func shouldSendToPeer(peerPubKey *ed25519.PublicKey, ackPeers []*ed25519.PublicKey) bool {
	for _, p := range ackPeers {
		if *p == *peerPubKey {
			return false
		}
	}
	return true
}

func (c *chainObj) EnqueueRequestAckMsg(msg *messages.RequestAckMsgIn) {
	c.requestAckPeerMsgPipe.In() <- msg
	c.chainMetrics.CountMessages()
}

func (c *chainObj) handleRequestAckPeerMsg(msg *messages.RequestAckMsgIn) {
	c.log.Debugf("handleRequestAckPeerMsg message received from peer %v, reqID: %s", msg.SenderPubKey.String(), msg.ReqID.Base58())
	c.offLedgerReqsAcksMutex.Lock()
	defer c.offLedgerReqsAcksMutex.Unlock()
	c.offLedgerReqsAcks[*msg.ReqID] = append(c.offLedgerReqsAcks[*msg.ReqID], msg.SenderPubKey)
	c.chainMetrics.CountRequestAckMessages()
	c.log.Debugf("handleRequestAckPeerMsg comleted: reqID: %s", msg.ReqID.Base58())
}

func (c *chainObj) EnqueueMissingRequestIDsMsg(msg *messages.MissingRequestIDsMsgIn) {
	c.missingRequestIDsPeerMsgPipe.In() <- msg
	c.chainMetrics.CountMessages()
}

func (c *chainObj) handleMissingRequestIDsMsg(msg *messages.MissingRequestIDsMsgIn) {
	c.log.Debugf("handleMissingRequestIDsMsg message received from peer %v, number of reqIDs: %v", msg.SenderPubKey.String(), len(msg.IDs))
	if !c.pullMissingRequestsFromCommittee {
		c.log.Warnf("handleMissingRequestIDsMsg ignored: pull from committee disabled")
		return
	}
	for _, reqID := range msg.IDs {
		c.log.Debugf("handleMissingRequestIDsMsg: finding reqID %s...", reqID.Base58())
		if req := c.mempool.GetRequest(reqID); req != nil {
			resultMsg := &messages.MissingRequestMsg{Request: req}
			c.chainPeers.SendMsgByPubKey(msg.SenderPubKey, peering.PeerMessageReceiverChain, chain.PeerMsgTypeMissingRequest, resultMsg.Bytes())
			c.log.Warnf("handleMissingRequestIDsMsg: reqID %s sent to %v.", reqID.Base58(), msg.SenderPubKey.String())
		} else {
			c.log.Warnf("handleMissingRequestIDsMsg: reqID %s not found.", reqID.Base58())
		}
	}
	c.log.Debugf("handleMissingRequestIDsMsg completed")
}

func (c *chainObj) EnqueueMissingRequestMsg(msg *messages.MissingRequestMsg) {
	c.missingRequestPeerMsgPipe.In() <- msg
	c.chainMetrics.CountMessages()
}

func (c *chainObj) handleMissingRequestMsg(msg *messages.MissingRequestMsg) {
	c.log.Debugf("handleMissingRequestMsg message received, reqID: %v", msg.Request.ID().Base58())
	if !c.pullMissingRequestsFromCommittee {
		c.log.Warnf("handleMissingRequestMsg ignored: pull from committee disabled")
		return
	}
	if c.consensus.ShouldReceiveMissingRequest(msg.Request) {
		c.mempool.ReceiveRequest(msg.Request)
		c.log.Warnf("handleMissingRequestMsg request with ID %v added to mempool", msg.Request.ID().Base58())
	} else {
		c.log.Warnf("handleMissingRequestMsg ignored: consensus denied the need of request with ID %v", msg.Request.ID().Base58())
	}
}

func (c *chainObj) EnqueueTimerTick(tick int) {
	c.timerTickMsgPipe.In() <- messages.TimerTick(tick)
	c.chainMetrics.CountMessages()
}

func (c *chainObj) handleTimerTick(msg messages.TimerTick) {
	if msg%2 == 0 {
		c.stateMgr.EnqueueTimerMsg(msg / 2)
	} else if c.consensus != nil {
		c.consensus.EnqueueTimerMsg(msg / 2)
	}
	if msg%40 == 0 {
		stats := c.mempool.Info()
		c.log.Debugf("mempool total = %d, ready = %d, in = %d, out = %d", stats.TotalPool, stats.ReadyCounter, stats.InPoolCounter, stats.OutPoolCounter)
	}
}
