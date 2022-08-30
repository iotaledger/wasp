// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import (
	"bytes"

	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
)

// MsgWrapper can be used to compose an algorithm out of other abstractions.
// These messages are meant to wrap and route the messages of the sub-algorithms.
type MsgWrapper struct {
	msgType       byte
	subsystemFunc func(subsystem byte, index int) (GPA, error) // Resolve a subsystem GPA based on its code and index.
}

func NewMsgWrapper(msgType byte, subsystemFunc func(subsystem byte, index int) (GPA, error)) *MsgWrapper {
	return &MsgWrapper{msgType, subsystemFunc}
}

func (w *MsgWrapper) WrapMessage(subsystem byte, index int, msg Message) Message {
	return &WrappingMsg{w.msgType, subsystem, index, msg}
}

func (w *MsgWrapper) WrapMessages(subsystem byte, index int, msgs OutMessages) OutMessages {
	if msgs == nil {
		return nil
	}
	wrapped := NoMessages()
	msgs.MustIterate(func(msg Message) {
		wrapped.Add(w.WrapMessage(subsystem, index, msg))
	})
	return wrapped
}

func (w *MsgWrapper) UnmarshalMessage(data []byte) (Message, error) {
	r := bytes.NewReader(data)
	msgType, err := util.ReadByte(r)
	if err != nil {
		return nil, xerrors.Errorf("cannot decode MsgWrapper::msgType: %w", msgType)
	}
	if msgType != w.msgType {
		return nil, xerrors.Errorf("invalid MsgWrapper::msgType, got %v, expected %v", msgType, w.msgType)
	}
	subsystem, err := util.ReadByte(r)
	if err != nil {
		return nil, err
	}
	var indexU16 uint16
	if err := util.ReadUint16(r, &indexU16); err != nil {
		return nil, err
	}
	index := int(indexU16)
	wrappedBin, err := util.ReadBytes16(r)
	if err != nil {
		return nil, err
	}

	subGPA, err := w.subsystemFunc(subsystem, index)
	if err != nil {
		return nil, err
	}
	wrapped, err := subGPA.UnmarshalMessage(wrappedBin)
	if err != nil {
		return nil, err
	}

	return &WrappingMsg{msgType, subsystem, index, wrapped}, nil
}

// The message that contains another, and its routing info.
type WrappingMsg struct {
	msgType   byte
	subsystem byte
	index     int
	wrapped   Message
}

var _ Message = &WrappingMsg{}

func (m *WrappingMsg) Subsystem() byte {
	return m.subsystem
}

func (m *WrappingMsg) Index() int {
	return m.index
}

func (m *WrappingMsg) Wrapped() Message {
	return m.wrapped
}

func (m *WrappingMsg) Recipient() NodeID {
	return m.wrapped.Recipient()
}

func (m *WrappingMsg) SetSender(sender NodeID) {
	m.wrapped.SetSender(sender)
}

func (m *WrappingMsg) MarshalBinary() ([]byte, error) {
	w := &bytes.Buffer{}
	if err := util.WriteByte(w, m.msgType); err != nil {
		return nil, err
	}
	if err := util.WriteByte(w, m.subsystem); err != nil {
		return nil, err
	}
	if err := util.WriteUint16(w, uint16(m.index)); err != nil {
		return nil, err
	}
	bin, err := m.wrapped.MarshalBinary()
	if err != nil {
		return nil, err
	}
	if err := util.WriteBytes16(w, bin); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}
