// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainimpl

import (
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

func init() {
	chain.RegisterChainConstructor(newCommitteeObj)
}

func (c *chainObj) IsOpenQueue() bool {
	if c.IsDismissed() {
		return false
	}
	if c.isOpenQueue.Load() {
		return true
	}
	c.mutexIsReady.Lock()
	defer c.mutexIsReady.Unlock()

	return c.checkReady()
}

func (c *chainObj) SetReadyStateManager() {
	if c.IsDismissed() {
		return
	}
	c.mutexIsReady.Lock()
	defer c.mutexIsReady.Unlock()

	c.isReadyStateManager = true
	c.log.Debugf("State Manager object was created")
	c.checkReady()
}

func (c *chainObj) SetReadyConsensus() {
	if c.IsDismissed() {
		return
	}
	c.mutexIsReady.Lock()
	defer c.mutexIsReady.Unlock()

	c.isReadyConsensus = true
	c.log.Debugf("consensus object was created")
	c.checkReady()
}

func (c *chainObj) SetConnectPeriodOver() {
	if c.IsDismissed() {
		return
	}
	c.mutexIsReady.Lock()
	defer c.mutexIsReady.Unlock()

	c.isConnectPeriodOver = true
	c.log.Debugf("connect period is over")
	c.checkReady()
}

func (c *chainObj) SetQuorumOfConnectionsReached() {
	if c.IsDismissed() {
		return
	}
	c.mutexIsReady.Lock()
	defer c.mutexIsReady.Unlock()

	c.isQuorumOfConnectionsReached = true
	c.log.Debugf("quorum of connections has been reached")
	c.checkReady()
}

func (c *chainObj) isReady() bool {
	return c.isReadyConsensus &&
		c.isReadyStateManager &&
		c.isConnectPeriodOver &&
		c.isQuorumOfConnectionsReached
}

func (c *chainObj) checkReady() bool {
	if c.IsDismissed() {
		panic("dismissed")
	}
	if c.isReady() {
		c.isOpenQueue.Store(true)
		c.startTimer()
		c.onActivation()

		c.log.Infof("committee now is fully initialized")
		publisher.Publish("active_committee", c.chainID.String())
	}
	return c.isReady()
}

func (c *chainObj) startTimer() {
	go func() {
		tick := 0
		for c.isOpenQueue.Load() {
			time.Sleep(chain.TimerTickPeriod)
			c.ReceiveMessage(chain.TimerTick(tick))
			tick++
		}
	}()
}

func (c *chainObj) Dismiss() {
	c.log.Infof("Dismiss committee for %s", c.chainID.String())

	c.dismissOnce.Do(func() {
		c.isOpenQueue.Store(false)
		c.dismissed.Store(true)

		close(c.chMsg)
		c.peers.Detach(c.peersAttachRef)
		c.peers.Close()

		c.stateMgr.Close()
		c.operator.Close()
	})

	publisher.Publish("dismissed_committee", c.chainID.String())
}

func (c *chainObj) IsDismissed() bool {
	return c.dismissed.Load()
}

func (c *chainObj) ID() *coretypes.ChainID {
	return &c.chainID
}

func (c *chainObj) Color() *balance.Color {
	return &c.color
}

func (c *chainObj) Address() address.Address {
	return address.Address(c.chainID)
}

func (c *chainObj) Size() uint16 {
	return c.size
}

func (c *chainObj) Quorum() uint16 {
	return c.quorum
}

func (c *chainObj) ReceiveMessage(msg interface{}) {
	if c.isOpenQueue.Load() {
		select {
		case c.chMsg <- msg:
		default:
			c.log.Warnf("ReceiveMessage with type '%T' failed. Retrying after %s", msg, chain.ReceiveMsgChannelRetryDelay)
			go func() {
				time.Sleep(chain.ReceiveMsgChannelRetryDelay)
				c.ReceiveMessage(msg)
			}()
		}
	}
}

// SendMsg sends message to peer by index. It can be both committee peer or access peer.
// TODO: [KP] Maybe we can use a broadcast instead of this?
func (c *chainObj) SendMsg(targetPeerIndex uint16, msgType byte, msgData []byte) error {
	if peer, ok := c.peers.OtherNodes()[targetPeerIndex]; ok {
		peer.SendMsg(&peering.PeerMessage{
			ChainID:     c.chainID,
			SenderIndex: c.ownIndex,
			MsgType:     msgType,
			MsgData:     msgData,
		})
		return nil
	}
	return fmt.Errorf("SendMsg: wrong peer index")
}

func (c *chainObj) SendMsgToCommitteePeers(msgType byte, msgData []byte, ts int64) uint16 {
	msg := &peering.PeerMessage{
		ChainID:     (coretypes.ChainID)(c.chainID),
		SenderIndex: c.ownIndex,
		Timestamp:   ts,
		MsgType:     msgType,
		MsgData:     msgData,
	}
	c.peers.Broadcast(msg, false)
	return uint16(len(c.peers.OtherNodes())) // TODO: [KP] Reconsider this, we cannot guaranty if they are actually sent.
}

// sends message to the peer seq[seqIndex]. If receives error, seqIndex = (seqIndex+1) % size and repeats
// if is not able to send message after size attempts, returns an error
// seqIndex is start seqIndex
// returned index is seqIndex of the successful send
func (c *chainObj) SendMsgInSequence(msgType byte, msgData []byte, seqIndex uint16, seq []uint16) (uint16, error) {
	if len(seq) != int(c.Size()) || seqIndex >= c.Size() || !util.ValidPermutation(seq) {
		return 0, fmt.Errorf("SendMsgInSequence: wrong params")
	}
	numAttempts := uint16(0)
	for ; numAttempts < c.Size(); seqIndex = (seqIndex + 1) % c.Size() {
		if seq[seqIndex] >= c.Size() {
			return 0, fmt.Errorf("SendMsgInSequence: wrong params")
		}
		if err := c.SendMsg(seq[seqIndex], msgType, msgData); err == nil {
			return seqIndex, nil
		}
		numAttempts++
	}
	return 0, fmt.Errorf("failed to send")
}

// returns true if peer is alive. Used by the operator to determine current leader
func (c *chainObj) IsAlivePeer(peerIndex uint16) bool {
	allNodes := c.peers.AllNodes()
	if int(peerIndex) >= len(allNodes) {
		return false
	}
	if peerIndex == c.ownIndex {
		return true
	}
	if allNodes[peerIndex] == nil {
		c.log.Panicf("c.peers[peerIndex] == nil. peerIndex: %d, ownIndex: %d", peerIndex, c.ownIndex)
	}
	return allNodes[peerIndex].IsAlive()
}

func (c *chainObj) OwnPeerIndex() uint16 {
	return c.ownIndex
}

func (c *chainObj) NumPeers() uint16 {
	return uint16(len(c.peers.AllNodes()))
}

// first N peers are committee peers, the rest are access peers in any
func (c *chainObj) committeePeers() map[uint16]peering.PeerSender {
	return c.peers.AllNodes()
}

func (c *chainObj) HasQuorum() bool {
	count := uint16(0)
	for _, peer := range c.committeePeers() {
		if peer == nil {
			count++
		} else {
			if peer.IsAlive() {
				count++
			}
		}
		if count >= c.quorum {
			return true
		}
	}
	return false
}

func (c *chainObj) PeerStatus() []*chain.PeerStatus {
	ret := make([]*chain.PeerStatus, 0)
	for i, peer := range c.committeePeers() {
		status := &chain.PeerStatus{
			Index:  int(i),
			IsSelf: peer == nil || peer.NetID() == c.netProvider.Self().NetID(),
		}
		if status.IsSelf {
			status.PeeringID = c.netProvider.Self().NetID()
			status.Connected = true
		} else {
			status.PeeringID = peer.NetID()
			status.Connected = peer.IsAlive()
		}
		ret = append(ret, status)
	}
	return ret
}

func (c *chainObj) BlobCache() coretypes.BlobCache {
	return c.blobProvider
}

func (c *chainObj) GetRequestProcessingStatus(reqID *coretypes.RequestID) chain.RequestProcessingStatus {
	if c.IsDismissed() {
		return chain.RequestProcessingStatusUnknown
	}
	if c.isCommitteeNode.Load() {
		if c.IsDismissed() {
			return chain.RequestProcessingStatusUnknown
		}
		if c.operator.IsRequestInBacklog(reqID) {
			return chain.RequestProcessingStatusBacklog
		}
	}
	processed, err := state.IsRequestCompleted(c.ID(), reqID)
	if err != nil || !processed {
		return chain.RequestProcessingStatusUnknown
	}
	return chain.RequestProcessingStatusCompleted
}

func (c *chainObj) Processors() *processors.ProcessorCache {
	return c.procset
}

func (c *chainObj) EventRequestProcessed() *events.Event {
	return c.eventRequestProcessed
}
