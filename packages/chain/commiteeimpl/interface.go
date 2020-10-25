package commiteeimpl

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/plugins/peering"
	"github.com/iotaledger/wasp/plugins/publisher"
	"time"
)

func init() {
	chain.ConstructorNew = newCommitteeObj
}

func (c *committeeObj) IsOpenQueue() bool {
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

func (c *committeeObj) SetReadyStateManager() {
	if c.IsDismissed() {
		return
	}
	c.mutexIsReady.Lock()
	defer c.mutexIsReady.Unlock()

	c.isReadyStateManager = true
	c.log.Debugf("State Manager object was created")
	c.checkReady()
}

func (c *committeeObj) SetReadyConsensus() {
	if c.IsDismissed() {
		return
	}
	c.mutexIsReady.Lock()
	defer c.mutexIsReady.Unlock()

	c.isReadyConsensus = true
	c.log.Debugf("consensus object was created")
	c.checkReady()
}

func (c *committeeObj) SetConnectPeriodOver() {
	if c.IsDismissed() {
		return
	}
	c.mutexIsReady.Lock()
	defer c.mutexIsReady.Unlock()

	c.isConnectPeriodOver = true
	c.log.Debugf("connect period is over")
	c.checkReady()
}

func (c *committeeObj) SetQuorumOfConnectionsReached() {
	if c.IsDismissed() {
		return
	}
	c.mutexIsReady.Lock()
	defer c.mutexIsReady.Unlock()

	c.isQuorumOfConnectionsReached = true
	c.log.Debugf("quorum of connections has been reached")
	c.checkReady()
}

func (c *committeeObj) isReady() bool {
	return c.isReadyConsensus &&
		c.isReadyStateManager &&
		c.isConnectPeriodOver &&
		c.isQuorumOfConnectionsReached
}

func (c *committeeObj) checkReady() bool {
	if c.IsDismissed() {
		panic("dismissed")
	}
	if c.isReady() {
		c.isOpenQueue.Store(true)
		c.startTimer()
		c.onActivation()

		c.log.Infof("committee now is fully initialized")
		publisher.Publish("active_committee", c.address.String())
	}
	return c.isReady()
}

func (c *committeeObj) startTimer() {
	go func() {
		tick := 0
		for c.isOpenQueue.Load() {
			time.Sleep(chain.TimerTickPeriod)
			c.ReceiveMessage(chain.TimerTick(tick))
			tick++
		}
	}()
}

func (c *committeeObj) Dismiss() {
	c.log.Infof("Dismiss committee for %s", c.address.String())

	c.dismissOnce.Do(func() {
		c.isOpenQueue.Store(false)
		c.dismissed.Store(true)

		close(c.chMsg)

		for _, pa := range c.peers {
			if pa != nil {
				peering.StopUsingPeer(pa.PeeringId())
			}
		}
	})

	publisher.Publish("dismissed_committee", c.address.String())
}

func (c *committeeObj) IsDismissed() bool {
	return c.dismissed.Load()
}

func (c *committeeObj) Address() *address.Address {
	return &c.address
}

func (c *committeeObj) OwnerAddress() *address.Address {
	return &c.ownerAddress
}

func (c *committeeObj) Color() *balance.Color {
	return &c.color
}

func (c *committeeObj) Size() uint16 {
	return c.size
}

func (c *committeeObj) Quorum() uint16 {
	return c.quorum
}

func (c *committeeObj) ReceiveMessage(msg interface{}) {
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

// sends message to peer by index. It can be both committee peer or access peer
func (c *committeeObj) SendMsg(targetPeerIndex uint16, msgType byte, msgData []byte) error {
	if int(targetPeerIndex) >= len(c.peers) {
		return fmt.Errorf("SendMsg: wrong peer index")
	}
	peer := c.peers[targetPeerIndex]
	if peer == nil {
		return fmt.Errorf("SendMsg: wrong peer")
	}
	msg := &peering.PeerMessage{
		ChainID:     (coretypes.ChainID)(c.address),
		SenderIndex: c.ownIndex,
		MsgType:     msgType,
		MsgData:     msgData,
	}
	return peer.SendMsg(msg)
}

func (c *committeeObj) SendMsgToCommitteePeers(msgType byte, msgData []byte, ts int64) uint16 {
	msg := &peering.PeerMessage{
		ChainID:     (coretypes.ChainID)(c.address),
		SenderIndex: c.ownIndex,
		MsgType:     msgType,
		MsgData:     msgData,
	}
	return peering.SendMsgToPeers(msg, ts, c.committeePeers()...)
}

// sends message to the peer seq[seqIndex]. If receives error, seqIndex = (seqIndex+1) % size and repeats
// if is not able to send message after size attempts, returns an error
// seqIndex is start seqIndex
// returned index is seqIndex of the successful send
func (c *committeeObj) SendMsgInSequence(msgType byte, msgData []byte, seqIndex uint16, seq []uint16) (uint16, error) {
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
func (c *committeeObj) IsAlivePeer(peerIndex uint16) bool {
	if int(peerIndex) >= len(c.peers) {
		return false
	}
	if peerIndex == c.ownIndex {
		return true
	}
	if c.peers[peerIndex] == nil {
		c.log.Panicf("c.peers[peerIndex] == nil. peerIndex: %d, ownIndex: %d", peerIndex, c.ownIndex)
	}
	return c.peers[peerIndex].IsAlive()
}

func (c *committeeObj) OwnPeerIndex() uint16 {
	return c.ownIndex
}

func (c *committeeObj) NumPeers() uint16 {
	return uint16(len(c.peers))
}

// first N peers are committee peers, the rest are access peers in any
func (c *committeeObj) committeePeers() []*peering.Peer {
	return c.peers[:c.size]
}

func (c *committeeObj) HasQuorum() bool {
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

func (c *committeeObj) PeerStatus() []*chain.PeerStatus {
	ret := make([]*chain.PeerStatus, 0)
	for i, peer := range c.committeePeers() {
		status := &chain.PeerStatus{
			Index:  i,
			IsSelf: peer == nil,
		}
		if status.IsSelf {
			status.PeeringID = peering.MyNetworkId()
			status.Connected = true
		} else {
			status.PeeringID = peer.PeeringId()
			status.Connected = peer.IsAlive()
		}
		ret = append(ret, status)
	}
	return ret
}

func (c *committeeObj) GetRequestProcessingStatus(reqId *coretypes.RequestID) chain.RequestProcessingStatus {
	if c.IsDismissed() {
		return chain.RequestProcessingStatusUnknown
	}
	if c.isCommitteeNode.Load() {
		if c.IsDismissed() {
			return chain.RequestProcessingStatusUnknown
		}
		if c.operator.IsRequestInBacklog(reqId) {
			return chain.RequestProcessingStatusBacklog
		}
	}
	processed, err := state.IsRequestCompleted(c.Address(), reqId)
	if err != nil || !processed {
		return chain.RequestProcessingStatusUnknown
	}
	return chain.RequestProcessingStatusCompleted
}
