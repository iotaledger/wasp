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

type MessagePayload interface {
	MsgType() MessageType
}

type TypedMessageIn[T MessagePayload] struct {
	Sender  NodeID
	Payload T
}

func NewMessageIn(sender NodeID, payload MessagePayload) TypedMessageIn[MessagePayload] {
	return TypedMessageIn[MessagePayload]{
		Sender:  sender,
		Payload: payload,
	}
}

type TypedMessageOut[T MessagePayload] struct {
	Recipient NodeID
	Payload   MessagePayload
}

func NewMessageOut(recipient NodeID, payload MessagePayload) TypedMessageOut[MessagePayload] {
	return TypedMessageOut[MessagePayload]{
		Recipient: recipient,
		Payload:   payload,
	}
}

type (
	MessageIn  = TypedMessageIn[MessagePayload]
	MessageOut = TypedMessageOut[MessagePayload]
)

func AsTypedMessageIn[T MessagePayload](msg MessageIn) TypedMessageIn[T] {
	return TypedMessageIn[T]{
		Sender:  msg.Sender,
		Payload: msg.Payload.(T),
	}
}

func NewPayloadIn[Payload any](sender NodeID, payload Payload) PayloadIn[Payload] {
	return PayloadIn[Payload]{
		Sender:  sender,
		Payload: payload,
	}
}

// PayloadIn is not full in message - it is just a payload value with sender.
//
// TODO: Revisit "Message" and "Payload" namings according to new gpa message design.
// For now I'm adding this just to be able to merge branches.
type PayloadIn[Payload any] struct {
	Sender  NodeID
	Payload Payload
}

func NewPayloadOut[Payload any](recipient NodeID, payload Payload) PayloadOut[Payload] {
	return PayloadOut[Payload]{
		Recipient: recipient,
		Payload:   payload,
	}
}

// PayloadOut is not full out message - it is just a payload value with recipient.
//
// TODO: Revisit "Message" and "Payload" namings according to new gpa message design.
// For now I'm adding this just to be able to merge branches.
type PayloadOut[Payload any] struct {
	Recipient NodeID
	Payload   Payload
}

type (
	Input  any
	Output any
)

// GPA is a generic interface for functional style distributed algorithms.
// GPA stands for Generic Pure Algorithm.
type GPA interface {
	Input(inp Input) []MessageOut
	Message(msg MessageIn) []MessageOut
	Output() Output
	StatusString() string // Status of the protocol as a string.
	UnmarshalPayload(data []byte) (MessagePayload, error)
}

type (
	PayloadAllocator map[MessageType]func() MessagePayload
	PayloadFallback  map[MessageType]func(data []byte) (MessagePayload, error)
)

func MarshalPayload(p MessagePayload) ([]byte, error) {
	e := bcs.NewBytesEncoder()
	e.WriteByte(p.MsgType())
	e.Encode(p)
	return e.Bytes(), e.Err()
}

func UnmarshalPayload(data []byte, mapper PayloadAllocator, fallback ...PayloadFallback) (MessagePayload, error) {
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

type Logger interface {
	LogWarnf(msg string, args ...any)
}
