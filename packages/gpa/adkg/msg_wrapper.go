// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package adkg

import "github.com/iotaledger/wasp/packages/gpa"

type msgWrapperSubsystem byte

const (
	msgWrapperACSS msgWrapperSubsystem = iota
	msgWrapperRBC
)

type msgWrapper struct {
	subsystem msgWrapperSubsystem
	index     int
	wrapped   gpa.Message
}

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

func (m *msgWrapper) MarshalBinary() ([]byte, error) {
	return nil, nil // TODO: Implement.
}

func (m *msgWrapper) Recipient() gpa.NodeID {
	return m.wrapped.Recipient()
}
