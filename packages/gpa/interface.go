// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// package gpa stands for generic pure (distributed) algorithm.
package gpa

import "encoding"

type NodeID string

type Message interface {
	encoding.BinaryMarshaler
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

func (msg *BasicMessage) SetSender(sender NodeID) {
	msg.sender = sender
}

func (msg *BasicMessage) Sender() NodeID {
	return msg.sender
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
