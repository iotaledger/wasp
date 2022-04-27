// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

// MsgWrapper can be used to compose an algorithm out of other abstractions.
// These messages are meant to wrap and route the messages of the sub-algorithms.
type MsgWrapper struct {
	subsystem byte
	index     int
	wrapped   Message
}

func WrapMessage(subsystem byte, index int, msg Message) Message {
	return &MsgWrapper{subsystem: subsystem, index: index, wrapped: msg}
}

func WrapMessages(subsystem byte, index int, msgs []Message) []Message {
	wrapped := make([]Message, len(msgs))
	for i := range msgs {
		wrapped[i] = WrapMessage(subsystem, index, msgs[i])
	}
	return wrapped
}

func (m *MsgWrapper) Subsystem() byte {
	return m.subsystem
}

func (m *MsgWrapper) Index() int {
	return m.index
}

func (m *MsgWrapper) Wrapped() Message {
	return m.wrapped
}

func (m *MsgWrapper) Recipient() NodeID {
	return m.wrapped.Recipient()
}

func (m *MsgWrapper) SetSender(sender NodeID) {
	m.wrapped.SetSender(sender)
}

func (m *MsgWrapper) MarshalBinary() ([]byte, error) {
	return nil, nil // TODO: Implement.
}
