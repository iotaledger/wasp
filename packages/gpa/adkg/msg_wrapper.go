// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package adkg

import "github.com/iotaledger/wasp/packages/gpa"

type msgWrapperSubsystem byte

const (
	msgWrapperACSS msgWrapperSubsystem = iota
	msgWrapperRBC
)

type msgWrapper struct { // TODO: Use the generic structure.
	subsystem msgWrapperSubsystem
	index     int
	wrapped   gpa.Message
}

var _ gpa.Message = &msgWrapper{}

func WrapMessage(subsystem msgWrapperSubsystem, index int, msg gpa.Message) gpa.Message {
	return &msgWrapper{subsystem: subsystem, index: index, wrapped: msg}
}

func WrapMessages(subsystem msgWrapperSubsystem, index int, msgs []gpa.Message) []gpa.Message {
	wrapped := make([]gpa.Message, len(msgs))
	for i := range msgs {
		wrapped[i] = WrapMessage(subsystem, index, msgs[i])
	}
	return wrapped
}

func (m *msgWrapper) Subsystem() msgWrapperSubsystem {
	return m.subsystem
}

func (m *msgWrapper) Recipient() gpa.NodeID {
	return m.wrapped.Recipient()
}

func (m *msgWrapper) SetSender(sender gpa.NodeID) {
	m.wrapped.SetSender(sender)
}

func (m *msgWrapper) MarshalBinary() ([]byte, error) {
	return nil, nil // TODO: Implement.
}
