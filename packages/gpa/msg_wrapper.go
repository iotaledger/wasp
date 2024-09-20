// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import (
	"bytes"
	"fmt"
	"io"

	"github.com/iotaledger/wasp/packages/util/bcs"
)

// MsgWrapper can be used to compose an algorithm out of other abstractions.
// These messages are meant to wrap and route the messages of the sub-algorithms.
type MsgWrapper struct {
	msgType       MessageType
	subsystemFunc func(subsystem byte, index int) (GPA, error) // Resolve a subsystem GPA based on its code and index.
}

func NewMsgWrapper(msgType MessageType, subsystemFunc func(subsystem byte, index int) (GPA, error)) *MsgWrapper {
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

func (w *MsgWrapper) DelegateInput(subsystem byte, index int, input Input) (GPA, OutMessages, error) {
	sub, err := w.subsystemFunc(subsystem, index)
	if err != nil {
		return nil, nil, err
	}
	return sub, w.WrapMessages(subsystem, index, sub.Input(input)), nil
}

func (w *MsgWrapper) DelegateMessage(msg *WrappingMsg) (GPA, OutMessages, error) {
	sub, err := w.subsystemFunc(msg.Subsystem(), msg.Index())
	if err != nil {
		return nil, nil, err
	}
	return sub, w.WrapMessages(msg.Subsystem(), msg.Index(), sub.Message(msg.Wrapped())), nil
}

func (w *MsgWrapper) MarshalMessage(msg Message) ([]byte, error) {
	wrapped := msg.(*WrappingMsg)

	e := bcs.NewBytesEncoder()
	e.WriteByte(wrapped.subsystem)
	e.WriteUint16(uint16(wrapped.index))

	subGPA, err := w.subsystemFunc(wrapped.subsystem, wrapped.index)
	if err != nil {
		return nil, fmt.Errorf("retrieving subsystem GPA %v/%v: %w", wrapped.subsystem, wrapped.index, err)
	}

	encodedMsg, err := subGPA.MarshalMessage(wrapped.wrapped)
	if err != nil {
		return nil, fmt.Errorf("marshalling wrapped message: subsystem %v index %v: %w", wrapped.subsystem, wrapped.index, err)
	}

	e.Write(encodedMsg)

	return e.Bytes(), e.Err()
}

func (w *MsgWrapper) UnmarshalMessage(data []byte) (Message, error) {
	r := bytes.NewReader(data)
	d := bcs.NewDecoder(r)

	subsystem := d.ReadByte()
	index := int(d.ReadUint16())

	subGPA, err := w.subsystemFunc(subsystem, index)
	if err != nil {
		return nil, fmt.Errorf("retrieving subsystem GPA %v/%v: %w", subsystem, index, err)
	}

	wrappedMsgBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	wrapped, err := subGPA.UnmarshalMessage(wrappedMsgBytes)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling wrapped message: subsystem %v index %v: %w", subsystem, index, err)
	}

	return &WrappingMsg{
		msgType:   w.msgType,
		subsystem: subsystem,
		index:     index,
		wrapped:   wrapped,
	}, nil
}

// The message that contains another, and its routing info.
type WrappingMsg struct {
	msgType   MessageType
	subsystem byte
	index     int
	wrapped   Message
}

var _ Message = new(WrappingMsg)

func (msg *WrappingMsg) Subsystem() byte {
	return msg.subsystem
}

func (msg *WrappingMsg) Index() int {
	return msg.index
}

func (msg *WrappingMsg) Wrapped() Message {
	return msg.wrapped
}

func (msg *WrappingMsg) Recipient() NodeID {
	return msg.wrapped.Recipient()
}

func (msg *WrappingMsg) SetSender(sender NodeID) {
	msg.wrapped.SetSender(sender)
}

func (msg *WrappingMsg) MarshalBCS(w io.Writer) error {
	encodedMsg, err := bcs.Marshal(&msg.wrapped)
	if err != nil {
		return err
	}

	enc := bcs.NewEncoder(w)
	enc.Encode(msg.subsystem)
	enc.Encode(uint16(msg.index))
	enc.Encode(encodedMsg)

	return enc.Err()
}
