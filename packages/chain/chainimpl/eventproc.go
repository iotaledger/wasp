// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainimpl

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/xerrors"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/committee"
	"github.com/iotaledger/wasp/packages/chain/consensus"
	"github.com/iotaledger/wasp/packages/chain/consensus/journal"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/tcrypto"
)

//nolint:gocyclo
func (c *chainObj) recvLoop() {
	dismissChainMsgChannel := c.dismissChainMsgPipe.Out()
	aliasOutputChannel := c.aliasOutputPipe.Out()
	offLedgerRequestMsgChannel := c.offLedgerRequestPeerMsgPipe.Out()
	missingRequestIDsMsgChannel := c.missingRequestIDsPeerMsgPipe.Out()
	missingRequestMsgChannel := c.missingRequestPeerMsgPipe.Out()
	timerTickMsgChannel := c.timerTickMsgPipe.Out()
	for {
		select {
		case msg, ok := <-dismissChainMsgChannel:
			if ok {
				c.log.Debugf("Chainimpl::recvLoop, handleDismissChain...")
				c.handleDismissChain(msg.(DismissChainMsg))
				c.log.Debugf("Chainimpl::recvLoop, handleDismissChain... Done")
			} else {
				dismissChainMsgChannel = nil
			}
		case msg, ok := <-aliasOutputChannel:
			if ok {
				c.log.Debugf("Chainimpl::recvLoop, handleAliasOutput...")
				c.handleAliasOutput(msg.(*isc.AliasOutputWithID))
				c.log.Debugf("Chainimpl::recvLoop, handleAliasOutput... Done")
			} else {
				aliasOutputChannel = nil
			}
		case msg, ok := <-offLedgerRequestMsgChannel:
			if ok {
				c.log.Debugf("Chainimpl::recvLoop, handleOffLedgerRequestMsg...")
				c.handleOffLedgerRequestMsg(msg.(*messages.OffLedgerRequestMsgIn))
				c.log.Debugf("Chainimpl::recvLoop, handleOffLedgerRequestMsg... Done")
			} else {
				offLedgerRequestMsgChannel = nil
			}
		case msg, ok := <-missingRequestIDsMsgChannel:
			if ok {
				c.log.Debugf("Chainimpl::recvLoop, handleMissingRequestIDsMsg...")
				c.handleMissingRequestIDsMsg(msg.(*messages.MissingRequestIDsMsgIn))
				c.log.Debugf("Chainimpl::recvLoop, handleMissingRequestIDsMsg... Done")
			} else {
				missingRequestIDsMsgChannel = nil
			}
		case msg, ok := <-missingRequestMsgChannel:
			if ok {
				c.log.Debugf("Chainimpl::recvLoop, handleMissingRequestMsg...")
				c.handleMissingRequestMsg(msg.(*messages.MissingRequestMsg))
				c.log.Debugf("Chainimpl::recvLoop, handleMissingRequestMsg... Done")
			} else {
				missingRequestMsgChannel = nil
			}
		case msg, ok := <-timerTickMsgChannel:
			if ok {
				c.log.Debugf("Chainimpl::recvLoop, handleTimerTick...")
				c.handleTimerTick(msg.(messages.TimerTick))
				c.log.Debugf("Chainimpl::recvLoop, handleTimerTick... Done")
			} else {
				timerTickMsgChannel = nil
			}
		}
		if dismissChainMsgChannel == nil &&
			aliasOutputChannel == nil &&
			offLedgerRequestMsgChannel == nil &&
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

func (c *chainObj) EnqueueAliasOutput(chainOutput *isc.AliasOutputWithID) {
	c.aliasOutputPipe.In() <- chainOutput
	c.chainMetrics.CountMessages()
}

// handleAliasOutput processes the only chain output which exists on the chain's address
// If necessary, it creates/changes/rotates committee object
func (c *chainObj) handleAliasOutput(msg *isc.AliasOutputWithID) {
	msgStateIndex := msg.GetStateIndex()
	c.log.Debugf("handleAliasOutput output received: state index %v, ID %v",
		msgStateIndex, isc.OID(msg.ID()))
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
				VirtualState:    nil, // governance rotation doesn't produce a new virtual state
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
		c.log.Debugf("handleAliasOutput: output %v passed to state manager", isc.OID(msg.ID()))
	}
}

func (c *chainObj) rotateCommitteeIfNeeded(anchorOutput *isc.AliasOutputWithID, currentCmt chain.Committee) (bool, error) {
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

func (c *chainObj) createCommitteeIfNeeded(anchorOutput *isc.AliasOutputWithID) (bool, error) {
	// check if I am in the committee
	stateControllerAddress := anchorOutput.GetStateAddress()
	dkShare, err := c.getChainDKShare(stateControllerAddress)
	if err != nil {
		if errors.Is(err, tcrypto.ErrDKShareNotFound) {
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
	cmtDKShare, err := c.dkShareRegistryProvider.LoadDKShare(addr)
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
	c.log.Debugf("createNewCommitteeAndConsensus: creating new consensus object for chainID=%+v, committee=%+v", c.chainID, cmt)
	cmtN := int(cmt.Size())
	cmtF := cmtN - int(dkShare.GetT())
	consensusJournal, err := journal.LoadConsensusJournal(*c.chainID, cmt.Address(), c.consensusJournalRegistryProvider, cmtN, cmtF, c.log)
	if err != nil {
		return xerrors.Errorf("cannot load consensus journal: %w", err)
	}
	c.consensus = consensus.New(
		c,
		c.mempool,
		cmt,
		cmtPeerGroup,
		c.pullMissingRequestsFromCommittee,
		c.chainMetrics,
		c.dssNode,
		consensusJournal,
		c.wal,
		c.nodeConn.PublishTransaction,
	)
	c.setCommittee(cmt)
	return nil
}

func (c *chainObj) EnqueueOffLedgerRequestMsg(msg *messages.OffLedgerRequestMsgIn) {
	c.offLedgerRequestPeerMsgPipe.In() <- msg
	c.chainMetrics.CountMessages()
}

// addToPeersHaveReq adds a peer to the list of known peers that have a given request, DOES NOT LOCK THE MUTEX
func (c *chainObj) addToPeersHaveReq(reqID isc.RequestID, peer *cryptolib.PublicKey) {
	if c.offLedgerPeersHaveReq[reqID] == nil {
		c.offLedgerPeersHaveReq[reqID] = make(map[cryptolib.PublicKeyKey]bool)
	}
	c.offLedgerPeersHaveReq[reqID][peer.AsKey()] = true
}

func (c *chainObj) handleOffLedgerRequestMsg(msg *messages.OffLedgerRequestMsgIn) {
	c.log.Debugf("handleOffLedgerRequestMsg message received from peer %v, reqID: %s", msg.SenderPubKey.String(), msg.Req.ID())

	if err := c.validateRequest(msg.Req); err != nil {
		// this means some node broadcasted an invalid request (bad chainID or signature)
		// TODO should the sender node be punished somehow?
		c.log.Errorf("handleOffLedgerRequestMsg message ignored: %v", err)
		return
	}
	c.log.Debugf("handleOffLedgerRequestMsg: request %s has been validated", msg.Req.ID())

	c.offLedgerPeersHaveReqMutex.Lock()
	c.addToPeersHaveReq(msg.Req.ID(), msg.SenderPubKey)
	c.offLedgerPeersHaveReqMutex.Unlock()

	if !c.mempool.ReceiveRequest(msg.Req) {
		c.log.Errorf("handleOffLedgerRequestMsg message ignored: mempool hasn't accepted request %s", msg.Req.ID())
		return
	}
	c.log.Debugf("handleOffLedgerRequestMsg: request %s added to mempool", msg.Req.ID())
	c.broadcastOffLedgerRequest(msg.Req)
	c.log.Debugf("handleOffLedgerRequestMsg: request %s broadcasted", msg.Req.ID())
}

func (c *chainObj) validateRequest(req isc.OffLedgerRequest) error {
	if !req.ChainID().Equals(c.ID()) {
		return fmt.Errorf("chainID mismatch")
	}
	return req.VerifySignature()
}

func (c *chainObj) broadcastOffLedgerRequest(req isc.OffLedgerRequest) {
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

	sendMessage := func(peersThatHaveTheRequest map[cryptolib.PublicKeyKey]bool) {
		peerPubKeys := getPeerPubKeys(c.offledgerBroadcastUpToNPeers)
		for _, peerPubKey := range peerPubKeys {
			if peersThatHaveTheRequest[peerPubKey.AsKey()] {
				// this peer already has the request
				c.log.Debugf("broadcastOffLedgerRequest: skipping send offledger request ID: %s to peerPubKey: %s.", req.ID(), peerPubKey.String())
				continue
			}
			c.log.Debugf("broadcastOffLedgerRequest: sending offledger request ID: reqID: %s, peerPubKey: %s", req.ID(), peerPubKey.String())
			c.chainPeers.SendMsgByPubKey(peerPubKey, peering.PeerMessageReceiverChain, chain.PeerMsgTypeOffLedgerRequest, msg.Bytes())
			c.addToPeersHaveReq(msg.Req.ID(), peerPubKey)
		}
	}

	ticker := time.NewTicker(c.offledgerBroadcastInterval)
	stopBroadcast := func() {
		ticker.Stop()
		c.offLedgerPeersHaveReqMutex.Lock()
		delete(c.offLedgerPeersHaveReq, req.ID())
		c.offLedgerPeersHaveReqMutex.Unlock()
	}

	go func() {
		defer stopBroadcast()
		for {
			<-ticker.C
			// check if processed (request already left the mempool)
			if !c.mempool.HasRequest(req.ID()) {
				return
			}

			shouldStop := func() bool {
				// deep copy the list of peers that have the request (otherwise we get a pointer to the map instead)
				c.offLedgerPeersHaveReqMutex.Lock()
				defer c.offLedgerPeersHaveReqMutex.Unlock()

				peersThatHaveTheRequest := c.offLedgerPeersHaveReq[req.ID()]
				if cmt != nil && len(peersThatHaveTheRequest) >= int(cmt.Size())-1 {
					// this node is part of the committee and the message has already been received by every other committee node
					return true
				}

				sendMessage(peersThatHaveTheRequest)
				return false
			}()

			if shouldStop {
				return
			}
		}
	}()
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
		c.log.Debugf("handleMissingRequestIDsMsg: finding reqID %s...", reqID.String())
		if req := c.mempool.GetRequest(reqID); req != nil {
			resultMsg := &messages.MissingRequestMsg{Request: req}
			c.chainPeers.SendMsgByPubKey(msg.SenderPubKey, peering.PeerMessageReceiverChain, chain.PeerMsgTypeMissingRequest, resultMsg.Bytes())
			c.log.Warnf("handleMissingRequestIDsMsg: reqID %s sent to %v.", reqID, msg.SenderPubKey.String())
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
