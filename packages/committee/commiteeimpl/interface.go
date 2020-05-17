package commiteeimpl

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/plugins/peering"
	"time"
)

func init() {
	committee.ConstructorNew = newCommitteeObj
}

// implements Committee interface

func (c *committeeObj) OpenQueue() {
	c.isOpenQueue.Store(true)
}

func (c *committeeObj) Dismiss() {
	c.isOpenQueue.Store(false)
	close(c.chMsg)

	for i, pa := range c.scdata.NodeLocations {
		if i != int(c.ownIndex) {
			peering.StopUsingPeer(pa)
		}
	}
}

func (c *committeeObj) Address() *address.Address {
	return &c.scdata.Address
}

func (c *committeeObj) Color() *balance.Color {
	return &c.scdata.Color
}

func (c *committeeObj) Size() uint16 {
	return uint16(len(c.scdata.NodeLocations))
}

func (c *committeeObj) ReceiveMessage(msg interface{}) {
	if c.isOpenQueue.Load() {
		c.chMsg <- msg
	}
}

// sends message to peer with index
func (c *committeeObj) SendMsg(targetPeerIndex uint16, msgType byte, msgData []byte) error {
	if targetPeerIndex == c.ownIndex || int(targetPeerIndex) >= len(c.peers) {
		return fmt.Errorf("SendMsg: wrong peer index")
	}
	peer := c.peers[targetPeerIndex]
	msg := &peering.PeerMessage{
		Address:     c.scdata.Address,
		SenderIndex: c.ownIndex,
		MsgType:     msgType,
		MsgData:     msgData,
	}
	return peer.SendMsg(msg)
}

func (c *committeeObj) SendMsgToPeers(msgType byte, msgData []byte) (uint16, time.Time) {
	msg := &peering.PeerMessage{
		Address:     c.scdata.Address,
		SenderIndex: c.ownIndex,
		MsgType:     msgType,
		MsgData:     msgData,
	}
	return peering.SendMsgToPeers(msg, c.peers...)
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
	if peerIndex == c.ownIndex {
		return true
	}
	if int(peerIndex) >= len(c.peers) {
		return false
	}
	ret, _ := c.peers[peerIndex].IsAlive()
	return ret
}

func (c *committeeObj) OwnPeerIndex() uint16 {
	return c.ownIndex
}

func (c *committeeObj) MetaData() *registry.SCMetaData {
	return c.scdata
}
