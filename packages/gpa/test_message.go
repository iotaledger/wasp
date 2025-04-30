// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

const msgTypeTest MessageType = 0xff

// TestMessage is just a message for test cases.
type TestMessage struct {
	recipient NodeID
	sender    NodeID
	ID        int
}

var _ Message = new(TestMessage)

func (msg *TestMessage) MsgType() MessageType {
	return msgTypeTest
}

func (msg *TestMessage) Recipient() NodeID {
	return msg.recipient
}

func (msg *TestMessage) SetSender(sender NodeID) {
	msg.sender = sender
}
