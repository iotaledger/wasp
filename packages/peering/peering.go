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
	"bytes"
	"errors"
	"hash/crc32"
	"io"
	"math/rand"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/txstream/chopper"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/mr-tron/base58"
	"golang.org/x/xerrors"
)

const (
	MsgTypeReserved  = byte(0)
	MsgTypeHandshake = byte(1)
	MsgTypeMsgChunk  = byte(2)

	// FirstUserMsgCode is the first committee message type.
	// All the equal and larger msg types are committee messages.
	// those with smaller are reserved by the package for heartbeat and handshake messages
	FirstUserMsgCode = byte(0x10)

	chunkMessageOverhead = 8 + 1
)

var crc32q *crc32.Table

func init() {
	crc32q = crc32.MakeTable(0xD5828281)
}

// PeeringID is relates peers in different nodes for a particular
// communication group. E.g. PeeringID identifies a committee in
// the consensus, etc.
type PeeringID [ledgerstate.AddressLength]byte

func RandomPeeringID(seed ...[]byte) PeeringID {
	var pid PeeringID
	_, _ = rand.Read(pid[:])
	return pid
}

func (pid *PeeringID) String() string {
	return base58.Encode(pid[:])
}

func (pid *PeeringID) Read(r io.Reader) error {
	if n, err := r.Read(pid[:]); err != nil || n != ledgerstate.AddressLength {
		return xerrors.Errorf("error while parsing PeeringID (err=%v)", err)
	}
	return nil
}

func (pid *PeeringID) Write(w io.Writer) error {
	if n, err := w.Write(pid[:]); err != nil || n != ledgerstate.AddressLength {
		return xerrors.Errorf("error while serializing PeeringID (err=%v)", err)
	}
	return nil
}

// NetworkProvider stands for the peer-to-peer network, as seen
// from the viewpoint of a single participant.
type NetworkProvider interface {
	Run(stopCh <-chan struct{})
	Self() PeerSender
	PeerGroup(peerAddrs []string) (GroupProvider, error)
	PeerDomain(peerAddrs []string) (PeerDomainProvider, error)
	Attach(peeringID *PeeringID, callback func(recv *RecvEvent)) interface{}
	Detach(attachID interface{})
	PeerByNetID(peerNetID string) (PeerSender, error)
	PeerByPubKey(peerPub *ed25519.PublicKey) (PeerSender, error)
	PeerStatus() []PeerStatusProvider
}

// TrustedNetworkManager is used maintain a configuration which peers are trusted.
// In a typical implementation, this interface should be implemented by the same
// struct, that implements the NetworkProvider. These implementations should interact,
// e.g. when we distrust some peer, all the connections to it should be cut immediately.
type TrustedNetworkManager interface {
	IsTrustedPeer(pubKey ed25519.PublicKey) error
	TrustPeer(pubKey ed25519.PublicKey, netID string) (*TrustedPeer, error)
	DistrustPeer(pubKey ed25519.PublicKey) (*TrustedPeer, error)
	TrustedPeers() ([]*TrustedPeer, error)
}

// TrustedPeer carries a peer information we use to trust it.
type TrustedPeer struct {
	PubKey ed25519.PublicKey
	NetID  string
}

func TrustedPeerFromBytes(buf []byte) (*TrustedPeer, error) {
	var err error
	r := bytes.NewBuffer(buf)
	tp := TrustedPeer{}
	var keyBytes []byte
	if keyBytes, err = util.ReadBytes16(r); err != nil {
		return nil, err
	}
	tp.PubKey, _, err = ed25519.PublicKeyFromBytes(keyBytes)
	if err != nil {
		return nil, err
	}
	if tp.NetID, err = util.ReadString16(r); err != nil {
		return nil, err
	}
	return &tp, nil
}

func (tp *TrustedPeer) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	if err := util.WriteBytes16(&buf, tp.PubKey.Bytes()); err != nil {
		return nil, err
	}
	if err := util.WriteString16(&buf, tp.NetID); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (tp *TrustedPeer) PubKeyBytes() ([]byte, error) {
	return tp.PubKey.Bytes(), nil
}

type PeerCollection interface {
	Attach(peeringID *PeeringID, callback func(recv *RecvEvent)) interface{}
	Detach(attachID interface{})
	Close()
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
	PeerIndexByNetID(peerNetID string) (uint16, error)
	SendMsgByIndex(peerIdx uint16, msg *PeerMessage)
	Broadcast(msg *PeerMessage, includingSelf bool)
	ExchangeRound(
		peers map[uint16]PeerSender,
		recvCh chan *RecvEvent,
		retryTimeout time.Duration,
		giveUpTimeout time.Duration,
		sendCB func(peerIdx uint16, peer PeerSender),
		recvCB func(recv *RecvEvent) (bool, error),
	) error
	AllNodes() map[uint16]PeerSender   // Returns all the nodes in the group.
	OtherNodes() map[uint16]PeerSender // Returns other nodes in the group (excluding Self).
	PeerCollection
}

// PeerDomainProvider implements unordered set of peers which can dynamically change
// All peers in the domain shares same peeringID. Each peer within domain is identified via its netID
type PeerDomainProvider interface {
	SendMsgByNetID(netID string, msg *PeerMessage)
	SendMsgToRandomPeers(upToNumPeers uint16, msg *PeerMessage)
	SendSimple(netID string, msgType byte, msgData []byte)
	SendMsgToRandomPeersSimple(upToNumPeers uint16, msgType byte, msgData []byte)
	ReshufflePeers(seedBytes ...[]byte)
	PeerCollection
}

// PeerSender represents an interface to some remote peer.
type PeerSender interface {

	// NetID identifies the peer.
	NetID() string

	// PubKey of the peer is only available, when it is
	// authenticated, therefore it can return nil, if pub
	// key is not known yet. You can call await before calling
	// this function to ensure the public key is already resolved.
	PubKey() *ed25519.PublicKey

	// SendMsg works in an asynchronous way, and therefore the
	// errors are not returned here.
	SendMsg(msg *PeerMessage)

	// IsAlive indicates, if there is a working connection with the peer.
	// It is always an approximate state.
	IsAlive() bool

	// Await for the connection to be established, handshaked, and the
	// public key resolved.
	Await(timeout time.Duration) error

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
	PubKey() *ed25519.PublicKey
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
	PeeringID   PeeringID
	SenderIndex uint16 // TODO: Only meaningful in a group, and when calculated by the client.
	SenderNetID string // TODO: Non persistent. Only used by PeeringDomain, filled by the receiver
	Timestamp   int64
	MsgType     byte
	MsgData     []byte
}

//nolint:gocritic
func NewPeerMessageFromBytes(buf []byte) (*PeerMessage, error) {
	var err error
	r := bytes.NewBuffer(buf)
	m := PeerMessage{}
	if err = util.ReadInt64(r, &m.Timestamp); err != nil {
		return nil, err
	}
	if m.MsgType, err = util.ReadByte(r); err != nil {
		return nil, err
	}
	switch m.MsgType {
	case MsgTypeReserved:
	case MsgTypeHandshake:
		if m.MsgData, err = util.ReadBytes32(r); err != nil {
			return nil, err
		}
	case MsgTypeMsgChunk:
		if m.MsgData, err = util.ReadBytes32(r); err != nil {
			return nil, err
		}
	default:
		if err = m.PeeringID.Read(r); err != nil {
			return nil, err
		}
		if err = util.ReadUint16(r, &m.SenderIndex); err != nil {
			return nil, err
		}
		if m.MsgData, err = util.ReadBytes32(r); err != nil {
			return nil, err
		}
		var checksumCalc uint32
		var checksumRead uint32
		checksumCalc = crc32.Checksum(m.MsgData, crc32q)
		if err = util.ReadUint32(r, &checksumRead); err != nil {
			return nil, err
		}
		if checksumCalc != checksumRead {
			return nil, errors.New("message_checksum_invalid")
		}
	}
	return &m, nil
}

// NewPeerMessageFromChunks can return nil, if there is not enough chunks to reconstruct the message.
func NewPeerMessageFromChunks(chunkBytes []byte, chunkSize int, msgChopper *chopper.Chopper) (*PeerMessage, error) {
	var err error
	var msgBytes []byte
	if msgBytes, err = msgChopper.IncomingChunk(chunkBytes, chunkSize, chunkMessageOverhead); err != nil {
		return nil, err
	}
	if msgBytes == nil {
		return nil, nil
	}
	return NewPeerMessageFromBytes(msgBytes)
}

func (m *PeerMessage) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	if err := util.WriteInt64(&buf, m.Timestamp); err != nil {
		return nil, err
	}
	if err := util.WriteByte(&buf, m.MsgType); err != nil {
		return nil, err
	}
	switch m.MsgType {
	case MsgTypeReserved:
	case MsgTypeHandshake:
		if err := util.WriteBytes32(&buf, m.MsgData); err != nil {
			return nil, err
		}
	case MsgTypeMsgChunk:
		if err := util.WriteBytes32(&buf, m.MsgData); err != nil {
			return nil, err
		}
	default:
		if err := m.PeeringID.Write(&buf); err != nil {
			return nil, err
		}
		if err := util.WriteUint16(&buf, m.SenderIndex); err != nil {
			return nil, err
		}
		if err := util.WriteBytes32(&buf, m.MsgData); err != nil {
			return nil, err
		}
		var checksumCalc uint32
		checksumCalc = crc32.Checksum(m.MsgData, crc32q)
		if err := util.WriteUint32(&buf, checksumCalc); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func (m *PeerMessage) ChunkedBytes(chunkSize int, msgChopper *chopper.Chopper) ([][]byte, error) {
	var err error
	var msgBytes []byte
	if msgBytes, err = m.Bytes(); err != nil {
		return nil, err
	}
	var choppedBytes [][]byte
	var chopped bool
	choppedBytes, chopped, err = msgChopper.ChopData(msgBytes, chunkSize, chunkMessageOverhead)
	if err != nil {
		return nil, err
	}
	if chopped {
		msgs := make([][]byte, len(choppedBytes))
		for i := range choppedBytes {
			chunkMsg := PeerMessage{
				Timestamp: m.Timestamp,
				MsgType:   MsgTypeMsgChunk,
				MsgData:   choppedBytes[i],
			}
			if msgs[i], err = chunkMsg.Bytes(); err != nil {
				return nil, err
			}
		}
		return msgs, nil
	}
	return [][]byte{msgBytes}, nil
}

func (m *PeerMessage) IsUserMessage() bool {
	return m.MsgType >= FirstUserMsgCode
}
