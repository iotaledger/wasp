package commiteeimpl

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/consensus"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/statemgr"
	"github.com/iotaledger/wasp/plugins/peering"
	"go.uber.org/atomic"
	"time"
)

const (
	useTimer        = true
	timerTickPeriod = 20 * time.Millisecond
)

type committeeObj struct {
	isOperational atomic.Bool
	peers         []*peering.Peer
	ownIndex      uint16
	scdata        *registry.SCData
	chMsg         chan interface{}
	stateMgr      *statemgr.StateManager
	operator      *consensus.Operator
}

func init() {
	committee.New = newCommitteeObj
}

func newCommitteeObj(scdata *registry.SCData) (committee.Committee, error) {
	ownIndex, ok := peering.FindOwnIndex(scdata.NodeLocations)
	if !ok {
		return nil, fmt.Errorf("not processed by this node sc addr: %s", scdata.Address.String())
	}
	dkshare, keyExists, err := registry.GetDKShare(scdata.Address)
	if err != nil {
		return nil, err
	}
	if !keyExists {
		return nil, fmt.Errorf("unkniwn key. sc addr = %s", scdata.Address.String())
	}
	err = fmt.Errorf("sc data inconsstent with key parameteres for sc addr %s", scdata.Address.String())
	if *scdata.Address != *dkshare.Address {
		return nil, err
	}
	if len(scdata.NodeLocations) != int(dkshare.N) {
		return nil, err
	}
	if dkshare.Index != ownIndex {
		return nil, err
	}

	ret := &committeeObj{
		chMsg:    make(chan interface{}, 10),
		scdata:   scdata,
		peers:    make([]*peering.Peer, 0, len(scdata.NodeLocations)),
		ownIndex: ownIndex,
	}
	for i, pa := range scdata.NodeLocations {
		if i != int(ownIndex) {
			ret.peers[i] = peering.UsePeer(pa)
		}
	}

	ret.stateMgr = statemgr.New(ret)
	ret.operator = consensus.NewOperator(ret, dkshare)

	go func() {
		for msg := range ret.chMsg {
			ret.dispatchMessage(msg)
		}
	}()

	if useTimer {
		go func() {
			tick := 0
			for {
				time.Sleep(timerTickPeriod)
				ret.ReceiveMessage(committee.TimerTick(tick))
			}
		}()
	}

	return ret, nil
}

// implements commtypes.Committee interface

func (c *committeeObj) SetOperational() {
	c.isOperational.Store(true)
}

func (c *committeeObj) Dismiss() {
	c.isOperational.Store(false)
	close(c.chMsg)

	for i, pa := range c.scdata.NodeLocations {
		if i != int(c.ownIndex) {
			peering.StopUsingPeer(pa)
		}
	}
}

func (c *committeeObj) Address() *address.Address {
	return c.scdata.Address
}

func (c *committeeObj) Color() *balance.Color {
	return c.scdata.Color
}

func (c *committeeObj) Size() uint16 {
	return uint16(len(c.scdata.NodeLocations))
}

func (c *committeeObj) ReceiveMessage(msg interface{}) {
	if c.isOperational.Load() {
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
		Address:     *c.scdata.Address,
		SenderIndex: c.ownIndex,
		MsgType:     msgType,
		MsgData:     msgData,
	}
	return peer.SendMsg(msg)
}

func (c *committeeObj) SendMsgToPeers(msgType byte, msgData []byte) (uint16, time.Time) {
	msg := &peering.PeerMessage{
		Address:     *c.scdata.Address,
		SenderIndex: c.ownIndex,
		MsgType:     msgType,
		MsgData:     msgData,
	}
	return peering.SendMsgToPeers(msg, c.peers...)
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
