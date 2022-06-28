// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainimpl

import (
	"errors"
	"fmt"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/committee"
	"github.com/iotaledger/wasp/packages/chain/consensus"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"golang.org/x/xerrors"
)

func (c *chainObj) recvLoop() {
	dismissChainMsgChannel := c.dismissChainMsgPipe.Out()
	aliasOutputChannel := c.aliasOutputPipe.Out()
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
		case msg, ok := <-aliasOutputChannel:
			if ok {
				c.handleAliasOutput(msg.(*iscp.AliasOutputWithID))
			} else {
				aliasOutputChannel = nil
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
			aliasOutputChannel == nil &&
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

func (c *chainObj) EnqueueAliasOutput(chainOutput *iscp.AliasOutputWithID) {
	c.aliasOutputPipe.In() <- chainOutput
	c.chainMetrics.CountMessages()
}

// handleAliasOutput processes the only chain output which exists on the chain's address
// If necessary, it creates/changes/rotates committee object
func (c *chainObj) handleAliasOutput(msg *iscp.AliasOutputWithID) {
	msgStateIndex := msg.GetStateIndex()
	c.log.Debugf("handleAliasOutput output received: state index %v, ID %v",
		msgStateIndex, iscp.OID(msg.ID()))
	commitment, err := state.L1CommitmentFromAliasOutput(msg.GetAliasOutput())
	if err != nil {
		c.log.Error(xerrors.Errorf("handleAliasOutput: parsing L1 commitment failed %w", err))
		return
	}
	c.log.Debugf("handleAliasOutput: L1 commitment is %s", commitment)

	var isRotation bool
	if (c.lastSeenOutputStateIndex == nil) || (*c.lastSeenOutputStateIndex <= msgStateIndex) {
		if c.lastSeenOutputStateIndex == nil {
			c.log.Debugf("handleAliasOutput: received initial state output")
		} else {
			c.log.Debugf("handleAliasOutput: received output, which is not older than the known one with index %v", *c.lastSeenOutputStateIndex)
		}
		cmt := c.getCommittee()
		if cmt != nil {
			isRotation, err = c.rotateCommitteeIfNeeded(msg, cmt)
			if err != nil {
				c.log.Errorf("handleAliasOutput: committee rotation failed: %v", err)
				return
			}
		} else {
			isRotation, err = c.createCommitteeIfNeeded(msg)
			if err != nil {
				c.log.Errorf("handleAliasOutput: committee creation failed: %v", err)
				return
			}
		}
		c.lastSeenOutputStateIndex = &msgStateIndex
		if msgStateIndex != 0 && isRotation {
			c.processChainTransition(&chain.ChainTransitionEventData{
				IsGovernance:    true,
				VirtualState:    c.lastSeenVirtualState,
				ChainOutput:     msg,
				OutputTimestamp: time.Now(),
			})
		}
	} else {
		isRotation = false
		c.log.Debugf("handleAliasOutput: received output, which is older than the known one with index %v; committee rotation/creation will not be performed", *c.lastSeenOutputStateIndex)
	}
	if msgStateIndex != 0 && isRotation {
		c.log.Debugf("handleAliasOutput: it is a rotation transition; skipping passing it to state manager")
	} else {
		c.stateMgr.EnqueueAliasOutput(msg)
		c.log.Debugf("handleAliasOutput: output %v passed to state manager", iscp.OID(msg.ID()))
	}
}

func (c *chainObj) rotateCommitteeIfNeeded(anchorOutput *iscp.AliasOutputWithID, currentCmt chain.Committee) (bool, error) {
	currentCmtAddress := currentCmt.Address()
	anchorOutputAddress := anchorOutput.GetStateAddress()
	if currentCmtAddress.Equal(anchorOutputAddress) {
		c.log.Debugf("rotateCommitteeIfNeeded rotation is not needed: committee address %s is not changed", currentCmtAddress)
		return false, nil
	}
	c.log.Debugf("rotateCommitteeIfNeeded rotation is needed: committee address is changed %s -> %s", currentCmtAddress, anchorOutputAddress)

	// rotation needed
	// close current in any case
	currentCmt.Close()
	c.consensus.Close()
	c.setCommittee(nil)
	c.log.Infof("rotateCommitteeIfNeeded: CLOSED COMMITTEE for the state address %s", currentCmtAddress)
	c.consensus = nil

	_, err := c.createCommitteeIfNeeded(anchorOutput)
	return true, err
}

func (c *chainObj) createCommitteeIfNeeded(anchorOutput *iscp.AliasOutputWithID) (bool, error) {
	// check if I am in the committee
	stateControllerAddress := anchorOutput.GetStateAddress()
	dkShare, err := c.getChainDKShare(stateControllerAddress)
	if err != nil {
		if errors.Is(err, registry.ErrDKShareNotFound) {
			c.log.Warnf("createCommitteeIfNeeded: DKShare not found, committee not created, node will not participate in consensus. Address: %s", stateControllerAddress)
			return false, nil
		}
		return false, xerrors.Errorf("createCommitteeIfNeeded: unable to load dkShare: %w", err)
	}
	// create new committee
	if err = c.createNewCommitteeAndConsensus(dkShare); err != nil {
		return true, xerrors.Errorf("createCommitteeIfNeeded: creating committee and consensus failed %w", err)
	}
	c.log.Infof("createCommitteeIfNeeded: CREATED COMMITTEE for the state address %s", stateControllerAddress)
	return true, nil
}

func (c *chainObj) getChainDKShare(addr iotago.Address) (tcrypto.DKShare, error) {
	//
	// just in case check if I am among committee nodes
	// should not happen
	selfPubKey := c.netProvider.Self().PubKey()
	cmtDKShare, err := c.dksProvider.LoadDKShare(addr)
	if err != nil {
		return nil, err
	}
	for _, pubKey := range cmtDKShare.GetNodePubKeys() {
		if pubKey.Equals(selfPubKey) {
			return cmtDKShare, nil
		}
	}
	return nil, xerrors.Errorf("createCommitteeIfNeeded: I am not among nodes of the committee record. Inconsistency")
}

func (c *chainObj) createNewCommitteeAndConsensus(dkShare tcrypto.DKShare) error {
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
		return xerrors.Errorf("createNewCommitteeAndConsensus: failed to create committee object for state address %s: %w", dkShare.GetAddress(), err)
	}
	attachID := cmtPeerGroup.Attach(peering.PeerMessageReceiverChain, c.receiveCommitteePeerMessages)
	c.detachFromCommitteePeerMessagesFun = func() {
		cmtPeerGroup.Detach(attachID)
	}
	c.log.Debugf("creating new consensus object...")
	c.consensus = consensus.New(c, c.mempool, cmt, cmtPeerGroup, c.nodeConn, c.pullMissingRequestsFromCommittee, c.chainMetrics, c.wal)
	c.setCommittee(cmt)
	return nil
}

func (c *chainObj) EnqueueOffLedgerRequestMsg(msg *messages.OffLedgerRequestMsgIn) {
	c.offLedgerRequestPeerMsgPipe.In() <- msg
	c.chainMetrics.CountMessages()
}

func (c *chainObj) handleOffLedgerRequestMsg(msg *messages.OffLedgerRequestMsgIn) {
	c.log.Debugf("handleOffLedgerRequestMsg message received from peer %v, reqID: %s", msg.SenderPubKey.AsString(), msg.Req.ID())
	c.sendRequestAcknowledgementMsg(msg.Req.ID(), msg.SenderPubKey)

	if err := c.validateRequest(msg.Req); err != nil {
		// this means some node broadcasted an invalid request (bad chainID or signature)
		// TODO should the sender node be punished somehow?
		c.log.Errorf("handleOffLedgerRequestMsg message ignored: %v", err)
		return
	}
	if !c.mempool.ReceiveRequest(msg.Req) {
		c.log.Errorf("handleOffLedgerRequestMsg message ignored: mempool hasn't accepted it")
		return
	}
	c.broadcastOffLedgerRequest(msg.Req)
	c.log.Debugf("handleOffLedgerRequestMsg message added to mempool and broadcasted: reqID: %s", msg.Req.ID().String())
}

func (c *chainObj) sendRequestAcknowledgementMsg(reqID iscp.RequestID, peerPubKey *cryptolib.PublicKey) {
	if peerPubKey == nil {
		return
	}
	c.log.Debugf("sendRequestAcknowledgementMsg: reqID: %s, peerID: %s", reqID, peerPubKey.AsString())
	msg := &messages.RequestAckMsg{ReqID: &reqID}
	c.chainPeers.SendMsgByPubKey(peerPubKey, peering.PeerMessageReceiverChain, chain.PeerMsgTypeRequestAck, msg.Bytes())
}

func (c *chainObj) validateRequest(req iscp.OffLedgerRequest) error {
	if !req.ChainID().Equals(c.ID()) {
		return fmt.Errorf("chainID mismatch")
	}
	return req.VerifySignature()
}

func (c *chainObj) broadcastOffLedgerRequest(req iscp.OffLedgerRequest) {
	c.log.Debugf("broadcastOffLedgerRequest: toNPeers: %d, reqID: %s", c.offledgerBroadcastUpToNPeers, req.ID())
	msg := &messages.OffLedgerRequestMsg{
		ChainID: c.chainID,
		Req:     req,
	}
	cmt := c.getCommittee()
	getPeerPubKeys := c.chainPeers.GetRandomOtherPeers

	if cmt != nil {
		getPeerPubKeys = cmt.GetRandomValidators
	}

	sendMessage := func(ackPeers []*cryptolib.PublicKey) {
		peerPubKeys := getPeerPubKeys(c.offledgerBroadcastUpToNPeers)
		for _, peerPubKey := range peerPubKeys {
			if shouldSendToPeer(peerPubKey, ackPeers) {
				c.log.Debugf("sending offledger request ID: reqID: %s, peerPubKey: %s", req.ID(), peerPubKey.AsString())
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
			ackPeers := c.offLedgerReqsAcks[req.ID()]
			c.offLedgerReqsAcksMutex.RUnlock()
			if cmt != nil && len(ackPeers) >= int(cmt.Size())-1 {
				// this node is part of the committee and the message has already been received by every other committee node
				return
			}
			sendMessage(ackPeers)
		}
	}()
}

func shouldSendToPeer(peerPubKey *cryptolib.PublicKey, ackPeers []*cryptolib.PublicKey) bool {
	for _, p := range ackPeers {
		if p.Equals(peerPubKey) {
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
	c.log.Debugf("handleRequestAckPeerMsg message received from peer %v, reqID: %s", msg.SenderPubKey.AsString(), msg.ReqID)
	c.offLedgerReqsAcksMutex.Lock()
	defer c.offLedgerReqsAcksMutex.Unlock()
	c.offLedgerReqsAcks[*msg.ReqID] = append(c.offLedgerReqsAcks[*msg.ReqID], msg.SenderPubKey)
	c.chainMetrics.CountRequestAckMessages()
	c.log.Debugf("handleRequestAckPeerMsg comleted: reqID: %s", msg.ReqID.String())
}

func (c *chainObj) EnqueueMissingRequestIDsMsg(msg *messages.MissingRequestIDsMsgIn) {
	c.missingRequestIDsPeerMsgPipe.In() <- msg
	c.chainMetrics.CountMessages()
}

func (c *chainObj) handleMissingRequestIDsMsg(msg *messages.MissingRequestIDsMsgIn) {
	c.log.Debugf("handleMissingRequestIDsMsg message received from peer %v, number of reqIDs: %v", msg.SenderPubKey.AsString(), len(msg.IDs))
	if !c.pullMissingRequestsFromCommittee {
		c.log.Warnf("handleMissingRequestIDsMsg ignored: pull from committee disabled")
		return
	}
	for _, reqID := range msg.IDs {
		c.log.Debugf("handleMissingRequestIDsMsg: finding reqID %s...", reqID.String())
		if req := c.mempool.GetRequest(reqID); req != nil {
			resultMsg := &messages.MissingRequestMsg{Request: req}
			c.chainPeers.SendMsgByPubKey(msg.SenderPubKey, peering.PeerMessageReceiverChain, chain.PeerMsgTypeMissingRequest, resultMsg.Bytes())
			c.log.Warnf("handleMissingRequestIDsMsg: reqID %s sent to %v.", reqID, msg.SenderPubKey.AsString())
		} else {
			c.log.Warnf("handleMissingRequestIDsMsg: reqID %s not found.", reqID.String())
		}
	}
	c.log.Debugf("handleMissingRequestIDsMsg completed")
}

func (c *chainObj) EnqueueMissingRequestMsg(msg *messages.MissingRequestMsg) {
	c.missingRequestPeerMsgPipe.In() <- msg
	c.chainMetrics.CountMessages()
}

func (c *chainObj) handleMissingRequestMsg(msg *messages.MissingRequestMsg) {
	c.log.Debugf("handleMissingRequestMsg message received, reqID: %v", msg.Request.ID().String())
	if !c.pullMissingRequestsFromCommittee {
		c.log.Warnf("handleMissingRequestMsg ignored: pull from committee disabled")
		return
	}
	if c.consensus.ShouldReceiveMissingRequest(msg.Request) {
		c.mempool.ReceiveRequest(msg.Request)
		c.log.Warnf("handleMissingRequestMsg request with ID %v added to mempool", msg.Request.ID().String())
	} else {
		c.log.Warnf("handleMissingRequestMsg ignored: consensus denied the need of request with ID %v", msg.Request.ID().String())
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
}
