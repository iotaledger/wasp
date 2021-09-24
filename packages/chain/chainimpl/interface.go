// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainimpl

import (
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/iscp/request"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

func (c *chainObj) ID() *iscp.ChainID {
	return &c.chainID
}

func (c *chainObj) GlobalStateSync() coreutil.ChainStateSync {
	return c.chainStateSync
}

func (c *chainObj) GetCommitteeInfo() *chain.CommitteeInfo {
	cmt := c.getCommittee()
	if cmt == nil {
		return nil
	}
	return &chain.CommitteeInfo{
		Address:       cmt.DKShare().Address,
		Size:          cmt.Size(),
		Quorum:        cmt.Quorum(),
		QuorumIsAlive: cmt.QuorumIsAlive(),
		PeerStatus:    cmt.PeerStatus(),
	}
}

func (c *chainObj) startTimer() {
	go func() {
		c.stateMgr.Ready().MustWait()
		tick := 0
		for !c.IsDismissed() {
			time.Sleep(chain.TimerTickPeriod)
			c.ReceiveMessage(messages.TimerTick(tick))
			tick++
		}
	}()
}

func (c *chainObj) Dismiss(reason string) {
	c.log.Infof("Dismiss chain. Reason: '%s'", reason)

	c.dismissOnce.Do(func() {
		c.dismissed.Store(true)

		c.chMsg.Close()

		c.mempool.Close()
		c.stateMgr.Close()
		cmt := c.getCommittee()
		if cmt != nil {
			cmt.Close()
		}
		if c.consensus != nil {
			c.consensus.Close()
		}
		c.eventRequestProcessed.DetachAll()
		c.eventChainTransition.DetachAll()
	})

	publisher.Publish("dismissed_chain", c.chainID.Base58())
}

func (c *chainObj) IsDismissed() bool {
	return c.dismissed.Load()
}

// ReceiveMessage accepts an incoming message asynchronously.
func (c *chainObj) ReceiveMessage(msg interface{}) {
	c.receiveMessage(msg)
}

func (c *chainObj) receiveMessage(msg interface{}) {
	if c.IsDismissed() {
		return
	}
	c.chMsg.In() <- msg
}

func shouldSendToPeer(peerID string, ackPeers []string) bool {
	for _, p := range ackPeers {
		if p == peerID {
			return false
		}
	}
	return true
}

func (c *chainObj) broadcastOffLedgerRequest(req *request.OffLedger) {
	c.log.Debugf("broadcastOffLedgerRequest: toNPeers: %d, reqID: %s", c.offledgerBroadcastUpToNPeers, req.ID().Base58())
	msgData := messages.NewOffLedgerRequestMsg(&c.chainID, req).Bytes()
	committee := c.getCommittee()
	getPeerIDs := (*c.peers).GetRandomPeers

	if committee != nil {
		getPeerIDs = committee.GetRandomValidators
	}

	sendMessage := func(ackPeers []string) {
		peerIDs := getPeerIDs(c.offledgerBroadcastUpToNPeers)
		for _, peerID := range peerIDs {
			if shouldSendToPeer(peerID, ackPeers) {
				c.log.Debugf("sending offledger request ID: reqID: %s, peerID: %s", req.ID().Base58(), peerID)
				(*c.peers).SendSimple(peerID, messages.MsgOffLedgerRequest, msgData)
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
			if committee != nil && len(ackPeers) >= int(committee.Size())-1 {
				// this node is part of the committee and the message has already been received by every other committee node
				return
			}
			sendMessage(ackPeers)
		}
	}()
}

func (c *chainObj) ReceiveOffLedgerRequest(req *request.OffLedger, senderNetID string) {
	c.log.Debugf("ReceiveOffLedgerRequest: reqID: %s, peerID: %s", req.ID().Base58(), senderNetID)
	c.sendRequestAcknowledgementMsg(req.ID(), senderNetID)
	if !c.mempool.ReceiveRequest(req) {
		return
	}
	c.log.Debugf("ReceiveOffLedgerRequest - added to mempool: reqID: %s, peerID: %s", req.ID().Base58(), senderNetID)
	c.broadcastOffLedgerRequest(req)
}

func (c *chainObj) sendRequestAcknowledgementMsg(reqID iscp.RequestID, peerID string) {
	c.log.Debugf("sendRequestAcknowledgementMsg: reqID: %s, peerID: %s", reqID.Base58(), peerID)
	if peerID == "" {
		return
	}
	msgData := messages.NewRequestAckMsg(reqID).Bytes()
	(*c.peers).SendSimple(peerID, messages.MsgRequestAck, msgData)
}

func (c *chainObj) ReceiveRequestAckMessage(reqID *iscp.RequestID, peerID string) {
	c.log.Debugf("ReceiveRequestAckMessage: reqID: %s, peerID: %s", reqID.Base58(), peerID)
	c.offLedgerReqsAcksMutex.Lock()
	defer c.offLedgerReqsAcksMutex.Unlock()
	c.offLedgerReqsAcks[*reqID] = append(c.offLedgerReqsAcks[*reqID], peerID)
}

// SendMissingRequestsToPeer sends the requested missing requests by a peer
func (c *chainObj) SendMissingRequestsToPeer(msg *messages.MissingRequestIDsMsg, peerID string) {
	for _, reqID := range msg.IDs {
		c.log.Debugf("Sending MissingRequestsToPeer: reqID: %s, peerID: %s", reqID.Base58(), peerID)
		if req := c.mempool.GetRequest(reqID); req != nil {
			msg := messages.NewMissingRequestMsg(req)
			(*c.peers).SendSimple(peerID, messages.MsgMissingRequest, msg.Bytes())
		}
	}
}

func (c *chainObj) ReceiveTransaction(tx *ledgerstate.Transaction) {
	c.log.Debugf("ReceiveTransaction: %s", tx.ID().Base58())
	reqs, err := request.OnLedgerFromTransaction(tx, c.chainID.AsAddress())
	if err != nil {
		c.log.Warnf("failed to parse transaction %s: %v", tx.ID().Base58(), err)
		return
	}
	for _, req := range reqs {
		c.ReceiveRequest(req)
	}
	if chainOut := transaction.GetAliasOutput(tx, c.chainID.AsAddress()); chainOut != nil {
		c.ReceiveState(chainOut, tx.Essence().Timestamp())
	}
}

func (c *chainObj) ReceiveRequest(req iscp.Request) {
	c.log.Debugf("ReceiveRequest: %s", req.ID())
	c.mempool.ReceiveRequests(req)
}

func (c *chainObj) ReceiveState(stateOutput *ledgerstate.AliasOutput, timestamp time.Time) {
	c.log.Debugf("ReceiveState #%d: outputID: %s, stateAddr: %s",
		stateOutput.GetStateIndex(), iscp.OID(stateOutput.ID()), stateOutput.GetStateAddress().Base58())
	c.ReceiveMessage(&messages.StateMsg{
		ChainOutput: stateOutput,
		Timestamp:   timestamp,
	})
}

func (c *chainObj) ReceiveInclusionState(txID ledgerstate.TransactionID, inclusionState ledgerstate.InclusionState) {
	c.ReceiveMessage(&messages.InclusionStateMsg{
		TxID:  txID,
		State: inclusionState,
	}) // TODO special entry point
}

func (c *chainObj) ReceiveOutput(output ledgerstate.Output) {
	c.stateMgr.EventOutputMsg(output)
}

func (c *chainObj) BlobCache() registry.BlobCache {
	return c.blobProvider
}

func (c *chainObj) GetRequestProcessingStatus(reqID iscp.RequestID) chain.RequestProcessingStatus {
	if c.IsDismissed() {
		return chain.RequestProcessingStatusUnknown
	}
	if c.consensus != nil {
		if c.mempool.HasRequest(reqID) {
			return chain.RequestProcessingStatusBacklog
		}
	}
	c.stateReader.SetBaseline()
	processed, err := blocklog.IsRequestProcessed(c.stateReader.KVStoreReader(), &reqID)
	if err != nil || !processed {
		return chain.RequestProcessingStatusUnknown
	}
	return chain.RequestProcessingStatusCompleted
}

func (c *chainObj) Processors() *processors.Cache {
	return c.procset
}

func (c *chainObj) EventRequestProcessed() *events.Event {
	return c.eventRequestProcessed
}

func (c *chainObj) RequestProcessed() *events.Event {
	return c.eventRequestProcessed
}

func (c *chainObj) ChainTransition() *events.Event {
	return c.eventChainTransition
}

func (c *chainObj) Events() chain.ChainEvents {
	return c
}

// GetStateReader returns a new copy of the optimistic state reader, with own baseline
func (c *chainObj) GetStateReader() state.OptimisticStateReader {
	return state.NewOptimisticStateReader(c.db, c.chainStateSync)
}

func (c *chainObj) Log() *logger.Logger {
	return c.log
}
