// Package peering provides an overlay network for communicating
// between nodes in a peer-to-peer style with low overhead
// encoding and persistent connections. The network provides only
// the asynchronous communication.
//
// It is intended to use for the committee consensus protocol.
//
package peering

// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"go.dedis.ch/kyber/v3"
)

const (
	// FirstCommitteeMsgCode is the first committee message type.
	// All the equal and larger msg types are committee messages.
	// those with smaller are reserved by the package for heartbeat and handshake messages
	FirstCommitteeMsgCode = byte(0x10) // TODO: Peering should be independent of committees.
)

// NetworkProvider stands for the peer-to-peer network, as seen
// from the viewpoint of a single participant.
type NetworkProvider interface {
	Run(stopCh <-chan struct{})
	Self() PeerSender
	Group(peerAddrs []string) (GroupProvider, error)
	Attach(chainID *coretypes.ChainID, callback func(recv *RecvEvent)) interface{}
	Detach(attachID interface{})
	PeerByLocation(peerLoc string) (PeerSender, error)
	PeerByPubKey(peerPub kyber.Point) (PeerSender, error)
	PeerStatus() []PeerStatusProvider
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
	Broadcast(msg *PeerMessage, includingSelf bool)
	AllNodes() map[uint16]PeerSender   // Returns all the nodes in the group.
	OtherNodes() map[uint16]PeerSender // Returns other nodes in the group (excluding Self).
	SendMsgByIndex(peerIdx uint16, msg *PeerMessage)
	Close()
}

// PeerSender represents an interface to some remote peer.
type PeerSender interface {
	Location() string // TODO: Rename to NetID
	PubKey() kyber.Point
	SendMsg(msg *PeerMessage)
	IsAlive() bool
	Close()
}

// PeerStatusProvider is used to access the current state of the network peer
// withouth allocating it (increading usage counters, etc). This interface
// overlaps with the PeerSender, and most probably they both will be implemented
// by the same object.
type PeerStatusProvider interface {
	Location() string
	PubKey() kyber.Point
	IsInbound() bool
	IsAlive() bool
	NumUsers() int
}

// RecvEvent stands for a received message along with
// the reference to its sender peer.
type RecvEvent struct {
	From PeerSender
	Msg  *PeerMessage
}

// PeerMessage is an envelope for all the messages exchanged via
// the peering module.
type PeerMessage struct {
	ChainID     coretypes.ChainID
	SenderIndex uint16
	Timestamp   int64
	MsgType     byte
	MsgData     []byte
}
