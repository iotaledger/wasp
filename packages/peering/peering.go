// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package peering provides an overlay network for communicating
// between nodes in a peer-to-peer style with low overhead
// encoding and persistent connections. The network provides only
// the asynchronous communication.
//
// It is intended to use for the committee consensus protocol.
//
package peering

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"golang.org/x/xerrors"
)

const (
	// FirstUserMsgCode is the first committee message type.
	// All the equal and larger msg types are committee messages.
	// those with smaller are reserved by the package for heartbeat and handshake messages
	FirstUserMsgCode                = byte(0x10)
	PeerMessageReceiverStateManager = byte(iota)
	PeerMessageReceiverConsensus
	PeerMessageReceiverCommonSubset
	PeerMessageReceiverChain
	PeerMessageReceiverChainDSS
	PeerMessageReceiverDkg
	PeerMessageReceiverDkgInit
)

// NetworkProvider stands for the peer-to-peer network, as seen
// from the viewpoint of a single participant.
type NetworkProvider interface {
	Run(ctx context.Context)
	Self() PeerSender
	PeerGroup(peeringID PeeringID, peerPubKeys []*cryptolib.PublicKey) (GroupProvider, error)
	PeerDomain(peeringID PeeringID, peerAddrs []*cryptolib.PublicKey) (PeerDomainProvider, error)
	PeerByPubKey(peerPub *cryptolib.PublicKey) (PeerSender, error)
	SendMsgByPubKey(pubKey *cryptolib.PublicKey, msg *PeerMessageData)
	PeerStatus() []PeerStatusProvider
	Attach(peeringID *PeeringID, receiver byte, callback func(recv *PeerMessageIn)) interface{}
	Detach(attachID interface{})
}

// TrustedNetworkManager is used maintain a configuration which peers are trusted.
// In a typical implementation, this interface should be implemented by the same
// struct, that implements the NetworkProvider. These implementations should interact,
// e.g. when we distrust some peer, all the connections to it should be cut immediately.
type TrustedNetworkManager interface {
	IsTrustedPeer(pubKey *cryptolib.PublicKey) error
	TrustPeer(pubKey *cryptolib.PublicKey, netID string) (*TrustedPeer, error)
	DistrustPeer(pubKey *cryptolib.PublicKey) (*TrustedPeer, error)
	TrustedPeers() ([]*TrustedPeer, error)
}

// GroupProvider stands for a subset of a peer-to-peer network
// that is responsible for achieving some common goal, eg,
// consensus committee, DKG group, etc.
//
// Indexes are only meaningful in the groups, not in the
// network or a particular peers.
type GroupProvider interface {
	SelfIndex() uint16
	PeerIndex(peer PeerSender) (uint16, error)
	PeerIndexByPubKey(peerPubKey *cryptolib.PublicKey) (uint16, error)
	PubKeyByIndex(index uint16) (*cryptolib.PublicKey, error)
	Attach(receiver byte, callback func(recv *PeerMessageGroupIn)) interface{}
	Detach(attachID interface{})
	SendMsgByIndex(peerIdx uint16, msgReceiver byte, msgType byte, msgData []byte)
	SendMsgBroadcast(msgReceiver byte, msgType byte, msgData []byte, except ...uint16)
	ExchangeRound(
		peers map[uint16]PeerSender,
		recvCh chan *PeerMessageIn,
		retryTimeout time.Duration,
		giveUpTimeout time.Duration,
		sendCB func(peerIdx uint16, peer PeerSender),
		recvCB func(recv *PeerMessageGroupIn) (bool, error),
	) error
	AllNodes(except ...uint16) map[uint16]PeerSender   // Returns all the nodes in the group except specified.
	OtherNodes(except ...uint16) map[uint16]PeerSender // Returns other nodes in the group (excluding Self and specified).
	Close()
}

// PeerDomainProvider implements unordered set of peers which can dynamically change
// All peers in the domain shares same peeringID. Each peer within domain is identified via its netID
type PeerDomainProvider interface {
	ReshufflePeers()
	GetRandomOtherPeers(upToNumPeers int) []*cryptolib.PublicKey
	UpdatePeers(newPeerPubKeys []*cryptolib.PublicKey)
	Attach(receiver byte, callback func(recv *PeerMessageIn)) interface{}
	Detach(attachID interface{})
	SendMsgByPubKey(pubKey *cryptolib.PublicKey, msgReceiver byte, msgType byte, msgData []byte)
	PeerStatus() []PeerStatusProvider
	Close()
}

// PeerSender represents an interface to some remote peer.
type PeerSender interface {
	// NetID identifies the peer.
	NetID() string

	// PubKey of the peer is only available, when it is
	// authenticated, therefore it can return nil, if pub
	// key is not known yet. You can call await before calling
	// this function to ensure the public key is already resolved.
	PubKey() *cryptolib.PublicKey

	// SendMsg works in an asynchronous way, and therefore the
	// errors are not returned here.
	SendMsg(msg *PeerMessageData)

	// IsAlive indicates, if there is a working connection with the peer.
	// It is always an approximate state.
	IsAlive() bool

	// Await for the connection to be established, handshaked, and the
	// public key resolved.
	Await(timeout time.Duration) error

	// Provides a read-only representation of this sender.
	Status() PeerStatusProvider

	// Close releases the reference to the peer, this informs the network
	// implementation, that it can disconnect, cleanup resources, etc.
	// You need to get new reference to the peer (PeerSender) to use it again.
	Close()
}

// PeerStatusProvider is used to access the current state of the network peer
// without allocating it (increading usage counters, etc). This interface
// overlaps with the PeerSender, and most probably they both will be implemented
// by the same object.
type PeerStatusProvider interface {
	NetID() string
	PubKey() *cryptolib.PublicKey
	IsAlive() bool
	NumUsers() int
}

// ParseNetID parses the NetID and returns the corresponding host and port.
func ParseNetID(netID string) (string, int, error) {
	parts := strings.Split(netID, ":")
	if len(parts) != 2 {
		return "", 0, xerrors.Errorf("invalid NetID: %v", netID)
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, xerrors.Errorf("invalid port in NetID: %v", netID)
	}
	return parts[0], port, nil
}
