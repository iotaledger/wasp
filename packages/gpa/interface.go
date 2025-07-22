// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package gpa stands for generic pure (distributed) algorithm.
package gpa

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"

	bcs "github.com/iotaledger/bcs-go"

	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/util"
)

type MessageType = byte

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
	return hexutil.Encode(niT[:])
}

func (niT NodeID) ShortString() string {
	return hexutil.Encode(niT[:4]) // 4 bytes - 8 hexadecimal digits
}

type Message interface {
	Recipient() NodeID // The sender should indicate the recipient.
	SetSender(NodeID)  // The transport later will set a validated sender for a message.
	MsgType() MessageType
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

// OutMessages is a buffer for collecting out messages.
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

// GPA is a generic interface for functional style distributed algorithms.
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

func MarshalMessage(msg Message) ([]byte, error) {
	e := bcs.NewBytesEncoder()
	e.WriteByte(msg.MsgType())
	e.Encode(msg)

	return e.Bytes(), e.Err()
}

func UnmarshalMessage(data []byte, mapper Mapper, fallback ...Fallback) (Message, error) {
	r := bytes.NewReader(data)

	msgType, err := bcs.UnmarshalStream[MessageType](r)
	if err != nil {
		return nil, err
	}

	allocator := mapper[msgType]
	if allocator != nil {
		msg := allocator()
		_, err := bcs.UnmarshalStreamInto(r, &msg)

		return msg, err
	}

	if len(fallback) == 0 {
		return nil, fmt.Errorf("unexpected message type %d", msgType)
	}
	if len(fallback) > 1 {
		return nil, fmt.Errorf("too many fallbacks specified: %d", len(fallback))
	}

	unmarshaler := fallback[0][msgType]
	if unmarshaler == nil {
		return nil, fmt.Errorf("unexpected message type %d", msgType)
	}

	return unmarshaler(data[1:])
}

func MarshalMessages(msgs []Message) ([][]byte, error) {
	msgsBytes := make([][]byte, len(msgs))
	var err error

	for i := range msgs {
		msgsBytes[i], err = MarshalMessage(msgs[i])
		if err != nil {
			return nil, fmt.Errorf("msgs[%d]: %w", i, err)
		}
	}

	return msgsBytes, nil
}

type Logger interface {
	LogWarnf(msg string, args ...any)
}
