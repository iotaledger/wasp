// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

// Just a message for test cases.
type TestMessage struct {
	recipient NodeID
	sender    NodeID
	ID        int
}

var _ Message = &TestMessage{}

func (m *TestMessage) Recipient() NodeID {
	return m.recipient
}

func (m *TestMessage) SetSender(sender NodeID) {
	m.sender = sender
}

func (m *TestMessage) MarshalBinary() ([]byte, error) {
	panic("not important")
}

func (m *TestMessage) UnmarshalBinary(data []byte) error {
	panic("not important")
}
