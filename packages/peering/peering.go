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
	"bytes"
	"time"

	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/chopper"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/util"
	"go.dedis.ch/kyber/v3"
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

// NetworkProvider stands for the peer-to-peer network, as seen
// from the viewpoint of a single participant.
type NetworkProvider interface {
	Run(stopCh <-chan struct{})
	Self() PeerSender
	Group(peerAddrs []string) (GroupProvider, error)
	Attach(chainID *coretypes.ChainID, callback func(recv *RecvEvent)) interface{}
	Detach(attachID interface{})
	PeerByNetID(peerNetID string) (PeerSender, error)
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
	PeerIndexByNetID(peerNetID string) (uint16, error)
	PeerIndexByPubKey(peerPub kyber.Point) (uint16, error)
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
	Attach(chainID *coretypes.ChainID, callback func(recv *RecvEvent)) interface{}
	Detach(attachID interface{})
	Close()
}

// PeerSender represents an interface to some remote peer.
type PeerSender interface {
	NetID() string
	PubKey() kyber.Point
	SendMsg(msg *PeerMessage)
	IsAlive() bool
	Close()
}

// PeerStatusProvider is used to access the current state of the network peer
// without allocating it (increading usage counters, etc). This interface
// overlaps with the PeerSender, and most probably they both will be implemented
// by the same object.
type PeerStatusProvider interface {
	NetID() string
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
	SenderIndex uint16 // TODO: Only meaningful in a group, and when calculated by the client.
	Timestamp   int64
	MsgType     byte
	MsgData     []byte
}

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
		m.ChainID = coretypes.NilChainID
		if err = m.ChainID.Read(r); err != nil {
			return nil, err
		}
		if err = util.ReadUint16(r, &m.SenderIndex); err != nil {
			return nil, err
		}
		if m.MsgData, err = util.ReadBytes32(r); err != nil {
			return nil, err
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
	var err error
	var buf bytes.Buffer
	if err = util.WriteInt64(&buf, m.Timestamp); err != nil {
		return nil, err
	}
	if err = util.WriteByte(&buf, m.MsgType); err != nil {
		return nil, err
	}
	switch m.MsgType {
	case MsgTypeReserved:
	case MsgTypeHandshake:
		if err = util.WriteBytes32(&buf, m.MsgData); err != nil {
			return nil, err
		}
	case MsgTypeMsgChunk:
		if err = util.WriteBytes32(&buf, m.MsgData); err != nil {
			return nil, err
		}
	default:
		if err = m.ChainID.Write(&buf); err != nil {
			return nil, err
		}
		if err = util.WriteUint16(&buf, m.SenderIndex); err != nil {
			return nil, err
		}
		if err = util.WriteBytes32(&buf, m.MsgData); err != nil {
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
