// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

const msgTypeTest MessageType = 0xff

// TestMessage is just a message for test cases.
type TestMessage struct {
	ID int
}

var _ MessagePayload = new(TestMessage)

func (msg *TestMessage) MsgType() MessageType {
	return msgTypeTest
}
