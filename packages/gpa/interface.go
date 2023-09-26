// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// package gpa stands for generic pure (distributed) algorithm.
package gpa

import (
	"errors"
	"io"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type MessageType rwutil.Kind

func (m *MessageType) Read(rr *rwutil.Reader) {
	*m = MessageType(rr.ReadKind())
}

func (m MessageType) ReadAndVerify(rr *rwutil.Reader) {
	rr.ReadKindAndVerify(rwutil.Kind(m))
}

func (m MessageType) Write(ww *rwutil.Writer) {
	ww.WriteKind(rwutil.Kind(m))
}

type NodeID [32]byte

var _ util.ShortStringable = NodeID{}

func NodeIDFromPublicKey(pubKey *cryptolib.PublicKey) NodeID {
	nodeID := NodeID{}
	copy(nodeID[:], pubKey.AsBytes())
	return nodeID
}

func NodeIDsFromPublicKeys(pubKeys []*cryptolib.PublicKey) []NodeID {
	ret := make([]NodeID, len(pubKeys))
	for i := range pubKeys {
		ret[i] = NodeIDFromPublicKey(pubKeys[i])
	}
	return ret
}

func (niT NodeID) Equals(other NodeID) bool {
	return niT == other
}

func (niT NodeID) String() string {
	return iotago.EncodeHex(niT[:])
}

func (niT NodeID) ShortString() string {
	return iotago.EncodeHex(niT[:4]) // 4 bytes - 8 hexadecimal digits
}

type Message interface {
	Read(r io.Reader) error
	Write(w io.Writer) error
	Recipient() NodeID // The sender should indicate the recipient.
	SetSender(NodeID)  // The transport later will set a validated sender for a message.
}

type BasicMessage struct {
	sender    NodeID
	recipient NodeID
}

func NewBasicMessage(recipient NodeID) BasicMessage {
	return BasicMessage{recipient: recipient}
}

func (msg *BasicMessage) Recipient() NodeID {
	return msg.recipient
}

func (msg *BasicMessage) Sender() NodeID {
	return msg.sender
}

func (msg *BasicMessage) SetSender(sender NodeID) {
	msg.sender = sender
}

type Input interface{}

type Output interface{}

// A buffer for collecting out messages.
// It is used to decrease array reallocations, if a slice would be used directly.
// Additionally, you can safely append to the OutMessages while you iterate over it.
// It should be implemented as a deep-list, allowing efficient appends and iterations.
type OutMessages interface {
	//
	// Add single message to the out messages.
	Add(msg Message) OutMessages
	//
	// Add several messages.
	AddMany(msgs []Message) OutMessages
	//
	// Add all the messages collected to other OutMessages.
	// The added OutMsgs object is marked done here.
	AddAll(msgs OutMessages) OutMessages
	//
	// Mark this instance as freezed, after this it cannot be appended.
	Done() OutMessages
	//
	// Returns a number of elements in the collection.
	Count() int
	//
	// Iterates over the collection, stops on first error.
	// Collection can be appended while iterating.
	Iterate(callback func(msg Message) error) error
	//
	// Iterated over the collection.
	// Collection can be appended while iterating.
	MustIterate(callback func(msg Message))
	//
	// Returns contents of the collection as an array of messages.
	AsArray() []Message
}

// Generic interface for functional style distributed algorithms.
// GPA stands for Generic Pure Algorithm.
type GPA interface {
	Input(inp Input) OutMessages     // Can return nil for NoMessages.
	Message(msg Message) OutMessages // Can return nil for NoMessages.
	Output() Output
	StatusString() string // Status of the protocol as a string.
	UnmarshalMessage(data []byte) (Message, error)
}

type (
	Mapper   map[MessageType]func() Message
	Fallback map[MessageType]func(data []byte) (Message, error)
)

func UnmarshalMessage(data []byte, mapper Mapper, fallback ...Fallback) (Message, error) {
	rr := rwutil.NewBytesReader(data)
	kind := rr.ReadKind()
	if rr.Err != nil {
		return nil, rr.Err
	}
	msgType := MessageType(kind)
	allocator := mapper[msgType]
	if allocator == nil {
		if len(fallback) == 1 {
			unmarshaler := fallback[0][msgType]
			if unmarshaler != nil {
				return unmarshaler(data)
			}
		}
		return nil, errors.New("cannot map kind to message")
	}
	msg := allocator()
	rr.PushBack().WriteKind(kind)
	rr.Read(msg)
	return msg, rr.Err
}

type Logger interface {
	Warnf(msg string, args ...any)
}
