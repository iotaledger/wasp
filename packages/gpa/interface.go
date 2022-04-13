// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// package gpa stands for generic pure (distributed) algorithm.
package gpa

import "encoding"

type NodeID string

type Message interface {
	encoding.BinaryMarshaler
	Recipient() NodeID
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
