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

func NewMessageIn[Payload any](sender NodeID, payload Payload) MessageIn[Payload] {
	return MessageIn[Payload]{
		Sender:  sender,
		Payload: payload,
	}
}

type TypedMessageOut[Payload any] struct {
	Recipient NodeID
	Payload   Payload
}

func NewMessageOut(recipient NodeID, payload any) MessageOut {
	return MessageOut{
		Recipient: recipient,
		Payload:   payload,
	}
}

func AsTypedMessageIn[Payload any](msg MessageIn[any]) MessageIn[Payload] {
	return MessageIn[Payload]{
		Sender:  msg.Sender,
		Payload: msg.Payload.(Payload),
	}
}

type MessageIn[Payload any] struct {
	Sender  NodeID
	Payload Payload
}

type MessageOut = TypedMessageOut[any]

type PayloadWithKey[Key, Payload any] struct {
	Key     Key
	Payload Payload
}

func AddKey[Key any](key Key, msgs []MessageOut) []MessageOut {
	ret := make([]MessageOut, len(msgs))
	for i, msg := range msgs {
		ret[i] = MessageOut{
			Recipient: msg.Recipient,
			Payload: PayloadWithKey[Key, any]{
				Key:     key,
				Payload: msg.Payload,
			},
		}
	}
	return ret
}

// TODO: Refactor or remove this before merge
type SubsystemPayload[Key, Payload any] struct {
	SubsystemID string
	Key         Key
	Payload     Payload
}

func AddSubsystemID[Key any](subsystemID string, key Key, msgs []MessageOut) []MessageOut {
	ret := make([]MessageOut, len(msgs))
	for i, msg := range msgs {
		ret[i] = MessageOut{
			Recipient: msg.Recipient,
			Payload: SubsystemPayload[Key, any]{
				SubsystemID: subsystemID,
				Key:         key,
				Payload:     msg.Payload,
			},
		}
	}
	return ret
}

type (
	Input  any
	Output any
)

// GPA is a generic interface for functional style distributed algorithms.
// GPA stands for Generic Pure Algorithm.
type GPA interface {
	Input(inp Input) []MessageOut
	Message(msg MessageIn[any]) []MessageOut
	Output() Output
	StatusString() string // Status of the protocol as a string.
	UnmarshalPayload(data []byte) (any, error)
	MarshalPayload(payload any) ([]byte, error)
}

type (
	PayloadAllocator map[MessageType]func() any
)

func MarshalPayload(msgType MessageType, payload any) ([]byte, error) {
	e := bcs.NewBytesEncoder()
	e.WriteByte(msgType)
	e.Encode(payload)
	return e.Bytes(), e.Err()
}

func UnmarshalPayload(data []byte, mapper PayloadAllocator) (any, error) {
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

	return nil, fmt.Errorf("unexpected message type %d", msgType)
}

type Logger interface {
	LogWarnf(msg string, args ...any)
}
