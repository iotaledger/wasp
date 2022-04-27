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

type Input interface{}

type Output interface{}

type GPA interface {
	Input(inp Input) []Message
	Message(msg Message) []Message
	Output() Output
}

// A convenience function to return from the Input or Message functions in GPA.
func NoMessages() []Message {
	return make([]Message, 0)
}
