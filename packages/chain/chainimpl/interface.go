// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainimpl

import (
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/iscp/request"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

func (c *chainObj) ID() *iscp.ChainID {
	return c.chainID
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
			c.EnqueueTimerTick(tick)
			tick++
			time.Sleep(chain.TimerTickPeriod)
		}
	}()
}

func (c *chainObj) Dismiss(reason string) {
	c.log.Infof("Dismiss chain. Reason: '%s'", reason)

	c.dismissOnce.Do(func() {
		c.dismissed.Store(true)

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
		c.chainPeers.Close()

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

func (c *chainObj) StateCandidateToStateManager(virtualState state.VirtualStateAccess, outputID ledgerstate.OutputID) {
	c.stateMgr.EnqueueStateCandidateMsg(virtualState, outputID)
}

func shouldSendToPeer(peerPubKey *ed25519.PublicKey, ackPeers []*ed25519.PublicKey) bool {
	for _, p := range ackPeers {
		if *p == *peerPubKey {
			return false
		}
	}
	return true
}

func (c *chainObj) broadcastOffLedgerRequest(req *request.OffLedger) {
	c.log.Debugf("broadcastOffLedgerRequest: toNPeers: %d, reqID: %s", c.offledgerBroadcastUpToNPeers, req.ID().Base58())
	msg := &messages.OffLedgerRequestMsg{
		ChainID: c.chainID,
		Req:     req,
	}
	committee := c.getCommittee()
	getPeerPubKeys := c.chainPeers.GetRandomPeers

	if committee != nil {
		getPeerPubKeys = committee.GetRandomValidators
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
			if committee != nil && len(ackPeers) >= int(committee.Size())-1 {
				// this node is part of the committee and the message has already been received by every other committee node
				return
			}
			sendMessage(ackPeers)
		}
	}()
}

func (c *chainObj) sendRequestAcknowledgementMsg(reqID iscp.RequestID, peerPubKey *ed25519.PublicKey) {
	if peerPubKey == nil {
		return
	}
	c.log.Debugf("sendRequestAcknowledgementMsg: reqID: %s, peerID: %s", reqID.Base58(), peerPubKey.String())
	msg := &messages.RequestAckMsg{ReqID: &reqID}
	c.chainPeers.SendMsgByPubKey(peerPubKey, peering.PeerMessageReceiverChain, chain.PeerMsgTypeRequestAck, msg.Bytes())
}

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

func (c *chainObj) ReceiveInclusionState(txID ledgerstate.TransactionID, inclusionState ledgerstate.InclusionState) {
	if c.consensus != nil {
		c.consensus.EnqueueInclusionsStateMsg(txID, inclusionState) // TODO special entry point
	}
}

func (c *chainObj) ReceiveOutput(output ledgerstate.Output) {
	c.stateMgr.EnqueueOutputMsg(output)
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

func (c *chainObj) GetChainNodes() []peering.PeerStatusProvider {
	return c.chainPeers.PeerStatus()
}

func (c *chainObj) GetCandidateNodes() []*governance.AccessNodeInfo {
	return c.candidateNodes
}

func (c *chainObj) Log() *logger.Logger {
	return c.log
}
