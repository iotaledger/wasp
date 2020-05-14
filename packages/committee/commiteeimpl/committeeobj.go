package commiteeimpl

import (
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/registry"
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
	scdata        *registry.SCMetaData
	chMsg         chan interface{}
	stateMgr      committee.StateManager
	operator      committee.Operator
	log           *logger.Logger
}

func init() {
	committee.New = newCommitteeObj
}

func newCommitteeObj(scdata *registry.SCMetaData, log *logger.Logger) (committee.Committee, error) {
	dkshare, keyExists, err := registry.GetDKShare(&scdata.Address)
	if err != nil {
		return nil, err
	}
	if !keyExists {
		return nil, fmt.Errorf("unkniwn key. sc addr = %s", scdata.Address.String())
	}
	err = fmt.Errorf("sc data inconsstent with key parameteres for sc addr %s", scdata.Address.String())
	if scdata.Address != *dkshare.Address {
		return nil, err
	}
	if err := checkNetworkLocations(scdata.NodeLocations, dkshare.N, dkshare.Index); err != nil {
		return nil, err
	}

	ret := &committeeObj{
		chMsg:    make(chan interface{}, 10),
		scdata:   scdata,
		peers:    make([]*peering.Peer, len(scdata.NodeLocations)),
		ownIndex: dkshare.Index,
		log:      log.Named("comm"),
	}
	myLocation := scdata.NodeLocations[dkshare.Index]
	for i, remoteLocation := range scdata.NodeLocations {
		if i != int(dkshare.Index) {
			ret.peers[i] = peering.UsePeer(remoteLocation, myLocation)
		}
	}

	//ret.stateMgr = statemgr.New(ret)
	//ret.operator = consensus.NewOperator(ret, dkshare)

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

// checkNetworkLocations checks if netLocations makes sense
func checkNetworkLocations(netLocations []string, n, index uint16) error {
	if len(netLocations) != int(n) {
		return fmt.Errorf("wrong number of network locations")
	}
	// check for duplicates
	for i := range netLocations {
		for j := i + 1; j < len(netLocations); j++ {
			if netLocations[i] == netLocations[j] {
				return errors.New("duplicate network locations in the list")
			}
		}
	}
	return peering.CheckMyNetworkID(netLocations[index])
}

// implements Committee interface

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
	return &c.scdata.Address
}

func (c *committeeObj) Color() *balance.Color {
	return &c.scdata.Color
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
