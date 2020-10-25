package peergroup

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/plugins/peering"
)

type PeerGroup struct {
	ownIndex      uint16
	chainID       coretypes.ChainID
	peers         []*peering.Peer
	iAmInTheGroup bool
}

// NewPeerGroup creates new group of peers. The chainid serves as prefix for messages
func NewPeerGroup(chainid coretypes.ChainID, netids []string) (*PeerGroup, error) {
	if util.ContainsDuplicates(netids) {
		return nil, fmt.Errorf("NewPeerGroup: contains netid duplicates")
	}
	ret := &PeerGroup{
		chainID: chainid,
		peers:   make([]*peering.Peer, len(netids)),
	}
	for i := range ret.peers {
		if peering.UsePeer(netids[i]) == nil {
			ret.ownIndex = (uint16)(i)
			ret.iAmInTheGroup = true
		}
	}
	return ret, nil
}

// Dismiss stops using all peers
func (pg *PeerGroup) Dismiss() {
	for _, peer := range pg.peers {
		if pg.peers != nil {
			peering.StopUsingPeer(peer.PeeringId())
		}
	}
}

// Size return number of peers
func (pg *PeerGroup) Size() uint16 {
	return uint16(len(pg.peers))
}

// SendMsg sends message to peer by index
func (pg *PeerGroup) SendMsg(targetPeerIndex uint16, msgType byte, msgData []byte) error {
	if targetPeerIndex >= pg.Size() || (pg.iAmInTheGroup && targetPeerIndex == pg.ownIndex) {
		return fmt.Errorf("PeerGroup.SendMsg: wrong peer index %d", targetPeerIndex)
	}
	peer := pg.peers[targetPeerIndex]
	msg := &peering.PeerMessage{
		ChainID:     pg.chainID,
		SenderIndex: pg.ownIndex,
		MsgType:     msgType,
		MsgData:     msgData,
	}
	return peer.SendMsg(msg)
}

func (pg *PeerGroup) SendMsgToPeers(msgType byte, msgData []byte, ts int64) uint16 {
	msg := &peering.PeerMessage{
		ChainID:     pg.chainID,
		SenderIndex: pg.ownIndex,
		MsgType:     msgType,
		MsgData:     msgData,
	}
	return peering.SendMsgToPeers(msg, ts, pg.peers...)
}

// returns true if peer is alive. Used by the operator to determine current leader
func (pg *PeerGroup) IsAlivePeer(peerIndex uint16) bool {
	if peerIndex >= pg.Size() {
		return false
	}
	if peerIndex == pg.ownIndex {
		return true
	}
	if pg.peers[peerIndex] == nil {
		panic("pg.peers[peerIndex] == nil")
	}
	return pg.peers[peerIndex].IsAlive()
}

func (pg *PeerGroup) OwnPeerIndex() uint16 {
	return pg.ownIndex
}
