package peering

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"go.dedis.ch/kyber/v3"
)

// NetworkProvider stands for the peer-to-peer network, as seen
// from the viewpoint of a single participant.
type NetworkProvider interface {
	Self() PeerSender
	Group(peerAddrs []string) (GroupProvider, error)
	Attach(chainID *coretypes.ChainID, callback func(recv *RecvEvent)) int
	Detach(attachID int)
	SendByLocation(peerLoc string, msg *PeerMessage)
}

// GroupProvider stands for a subset of a peer-to-peer network
// that is responsible for achieving some common goal, eg,
// consensus committee, DKG group, etc.
//
// Indexes are only meaningful in the groups, not in the
// network or a particular peers.
type GroupProvider interface {
	PeerIndex(peer PeerSender) (uint16, error)
	PeerIndexByLoc(peerLoc string) (uint16, error)
	PeerIndexByPub(peerPub kyber.Point) (uint16, error)
	Broadcast(msg *PeerMessage)
	AllNodes() map[uint16]PeerSender   // Returns all the nodes in the group.
	OtherNodes() map[uint16]PeerSender // Returns other nodes in the group (excluding Self).
	SendMsgByIndex(peerIdx uint16, msg *PeerMessage)
	Close()
}

// PeerSender represents an interface to some remote peer.
type PeerSender interface {
	Location() string
	PubKey() kyber.Point
	SendMsg(msg *PeerMessage)
}

// RecvEvent stands for a received message along with
// the reference to its sender peer.
type RecvEvent struct {
	From PeerSender
	Msg  *PeerMessage
}
