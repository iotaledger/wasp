// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acss

import "github.com/iotaledger/wasp/packages/gpa"

const (
	msgWrapperRBC byte = iota
)

type msgWrapper struct {
	subsystem byte
	wrapped   gpa.Message
}

func WrapMessages(subsystem byte, msgs []gpa.Message) []gpa.Message {
	wrapped := make([]gpa.Message, len(msgs))
	for i := range msgs {
		wrapped[i] = &msgWrapper{subsystem: subsystem, wrapped: msgs[i]}
	}
	return wrapped
}

func (m *msgWrapper) Subsystem() byte {
	return m.subsystem
}

func (m *msgWrapper) MarshalBinary() ([]byte, error) {
	return nil, nil // TODO: Implement.
}

func (m *msgWrapper) Recipient() gpa.NodeID {
	return m.wrapped.Recipient()
}
