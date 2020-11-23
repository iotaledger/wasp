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
	Attach(chainID coretypes.ChainID, callback func(from PeerSender, msg *PeerMessage))
	SendByLocation(peerLoc string, msg *PeerMessage)
}

// GroupProvider stands for a subset of a peer-to-peer network
// that is responsible for achieving some common goal, eg,
// consensus committee, DKG group, etc.
//
// Indexes are only meaningful in the groups, not in the
// network or a particular peers.
type GroupProvider interface {
	PeerIndex(peer PeerSender) (int, error)
	PeerIndexByLoc(peerLoc string) (int, error)
	PeerIndexByPub(peerPub kyber.Point) (int, error)
	// PeerByIndex(peerIdx int) PeerSender
	Broadcast(msg *PeerMessage)
	AllNodes() map[int]PeerSender   // Returns all the nodes in the group.
	OtherNodes() map[int]PeerSender // Returns other nodes in the group (excluding Self).
	SendMsgByIndex(peerIdx int, msg *PeerMessage)
	// Send(peerLoc string, msg *PeerMessage)
}

// PeerSender represents an interface to some remote peer.
// TODO: Consider other name, Peer is already taken.
type PeerSender interface {
	Location() string
	PubKey() kyber.Point
	SendMsg(msg *PeerMessage)
}
