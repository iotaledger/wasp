// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package peering provides an overlay network for communicating
// between nodes in a peer-to-peer style with low overhead
// encoding and persistent connections. The network provides only
// the asynchronous communication.
//
// It is intended to use for the committee consensus protocol.
package peering

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/v2/packages/cryptolib"
)

const (
	// FirstUserMsgCode is the first committee message type.
	// All the equal and larger msg types are committee messages.
	// those with smaller are reserved by the package for heartbeat and handshake messages
	FirstUserMsgCode     = byte(0x10)
	ReceiverStateManager = byte(iota)
	ReceiverConsensus
	ReceiverCommonSubset
	ReceiverChain
	ReceiverChainDSS
	ReceiverChainCons
	ReceiverDkg
	ReceiverDkgInit
	ReceiverMempool
	ReceiverAccessMgr
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
	Attach(peeringID *PeeringID, receiver byte, callback func(recv *PeerMessageIn)) context.CancelFunc
}

// TrustedNetworkManager is used maintain a configuration which peers are trusted.
// In a typical implementation, this interface should be implemented by the same
// struct, that implements the NetworkProvider. These implementations should interact,
// e.g. when we distrust some peer, all the connections to it should be cut immediately.
type TrustedNetworkManager interface {
	IsTrustedPeer(pubKey *cryptolib.PublicKey) error
	TrustPeer(name string, pubKey *cryptolib.PublicKey, peeringURL string) (*TrustedPeer, error)
	DistrustPeer(pubKey *cryptolib.PublicKey) (*TrustedPeer, error)
	TrustedPeers() ([]*TrustedPeer, error)
	TrustedPeersByPubKeyOrName(pubKeysOrNames []string) ([]*TrustedPeer, error)
	// The following has to register a callback receiving updates to a set of trusted peers.
	// Upon subscription the initial set of peers has to be passed without waiting for updates.
	// The function returns a cancel func.The context is used to cancel the subscription.
	TrustedPeersListener(callback func([]*TrustedPeer)) context.CancelFunc
}

// ValidateTrustedPeerParams performs basic checks on trusted peer parameters, to be used in all the implementations.
func ValidateTrustedPeerParams(name string, pubKey *cryptolib.PublicKey, peeringURL string) error {
	if name != pubKey.String() && strings.HasPrefix(name, "0x") {
		return fmt.Errorf("name cannot start with '0x' unless it is equal to pubKey")
	}
	if name == "" {
		return errors.New("name is mandatory for a trusted peer")
	}
	return nil
}

// QueryByPubKeyOrName resolves pubKeysOrNames to TrustedPeers. Fails if any of the names/keys cannot be resolved.
func QueryByPubKeyOrName(trustedPeers []*TrustedPeer, pubKeysOrNames []string) ([]*TrustedPeer, error) {
	result := make([]*TrustedPeer, len(pubKeysOrNames))
	for i, pubKeyOrName := range pubKeysOrNames {
		isPubKey := strings.HasPrefix(pubKeyOrName, "0x")
		var pubKey *cryptolib.PublicKey
		var err error
		if isPubKey {
			pubKey, err = cryptolib.PublicKeyFromString(pubKeyOrName)
			if err != nil {
				return nil, fmt.Errorf("cannot parse %v as pubKey: %w", pubKeyOrName, err)
			}
		}
		if isPubKey {
			peer, ok := lo.Find(trustedPeers, func(p *TrustedPeer) bool {
				return pubKey.Equals(p.PubKey())
			})
			if !ok {
				return nil, fmt.Errorf("cannot find trusted peer by pubKey=%v", pubKeyOrName)
			}
			result[i] = peer
		} else {
			peer, ok := lo.Find(trustedPeers, func(p *TrustedPeer) bool {
				return p.Name == pubKeyOrName
			})
			if !ok {
				return nil, fmt.Errorf("cannot find trusted peer by name=%v", pubKeyOrName)
			}
			result[i] = peer
		}
	}
	return result, nil
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
	Attach(receiver byte, callback func(recv *PeerMessageGroupIn)) context.CancelFunc
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
// All peers in the domain shares same peeringID. Each peer within domain is identified via its peeringURL
type PeerDomainProvider interface {
	ReshufflePeers()
	GetRandomOtherPeers(upToNumPeers int) []*cryptolib.PublicKey
	UpdatePeers(newPeerPubKeys []*cryptolib.PublicKey)
	Attach(receiver byte, callback func(recv *PeerMessageIn)) context.CancelFunc
	SendMsgByPubKey(pubKey *cryptolib.PublicKey, msgReceiver byte, msgType byte, msgData []byte)
	PeerStatus() []PeerStatusProvider
	Close()
}

// PeerSender represents an interface to some remote peer.
type PeerSender interface {
	// PeeringURL identifies the peer.
	PeeringURL() string

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
// without allocating it (increasing usage counters, etc). This interface
// overlaps with the PeerSender, and most probably they both will be implemented
// by the same object.
type PeerStatusProvider interface {
	Name() string
	PeeringURL() string
	PubKey() *cryptolib.PublicKey
	IsAlive() bool
	NumUsers() int
}

// ParsePeeringURL parses the peeringURL and returns the corresponding host and port.
func ParsePeeringURL(url string) (string, int, error) {
	parts := strings.Split(url, ":")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid peeringURL: %v", url)
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, fmt.Errorf("invalid port in peeringURL: %v", url)
	}
	return parts[0], port, nil
}
