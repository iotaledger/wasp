// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import (
	"fmt"

	bcs "github.com/iotaledger/bcs-go"
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

func (w *MsgWrapper) UnmarshalMessage(data []byte) (Message, error) {
	rawMsg, err := bcs.Unmarshal[rawWrappingMsg](data)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling wrapping msg: %w", err)
	}

	subGPA, err := w.subsystemFunc(rawMsg.Subsystem, rawMsg.Index)
	if err != nil {
		return nil, fmt.Errorf("retrieving subsystem GPA %v/%v: %w", rawMsg.Subsystem, rawMsg.Index, err)
	}

	wrapped, err := subGPA.UnmarshalMessage(rawMsg.WrappedMsgBytes)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling wrapped message: subsystem %v index %v: %w", rawMsg.Subsystem, rawMsg.Index, err)
	}

	return &WrappingMsg{
		msgType:   w.msgType,
		subsystem: rawMsg.Subsystem,
		index:     rawMsg.Index,
		wrapped:   wrapped,
	}, nil
}

// WrappingMsg is the message that contains another, and its routing info.
type WrappingMsg struct {
	msgType   MessageType
	subsystem byte
	index     int
	wrapped   Message
}

var _ Message = new(WrappingMsg)

func (msg *WrappingMsg) MsgType() MessageType {
	return msg.msgType
}

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

func (msg *WrappingMsg) MarshalBCS(e *bcs.Encoder) error {
	wrappedMsgBytes, err := MarshalMessage(msg.wrapped)
	if err != nil {
		return fmt.Errorf("marshaling wrapped message: %w", err)
	}

	e.Encode(rawWrappingMsg{
		Subsystem:       msg.subsystem,
		Index:           msg.index,
		WrappedMsgBytes: wrappedMsgBytes,
	})

	return nil
}

type rawWrappingMsg struct {
	Subsystem       byte
	Index           int `bcs:"type=u16"`
	WrappedMsgBytes []byte
}
