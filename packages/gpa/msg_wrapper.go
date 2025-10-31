// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import (
	"fmt"

	"github.com/samber/lo"

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

func (w *MsgWrapper) WrapMessageOut(subsystem byte, index int, msg *MessageOut) *MessageOut {
	return NewMessageOut(
		msg.Recipient,
		&WrappingMsg{w.msgType, subsystem, index, msg.Payload},
	)
}

func (w *MsgWrapper) WrapMessagesOut(subsystem byte, index int, msgs []*MessageOut) []*MessageOut {
	return lo.Map(msgs, func(msg *MessageOut, _ int) *MessageOut {
		return w.WrapMessageOut(subsystem, index, msg)
	})
}

func (w *MsgWrapper) WrapMessageIn(subsystem byte, index int, msg *MessageIn) *MessageIn {
	return NewMessageIn(
		msg.Sender,
		&WrappingMsg{w.msgType, subsystem, index, msg.Payload},
	)
}

func (w *MsgWrapper) WrapMessagesIn(subsystem byte, index int, msgs []*MessageIn) []*MessageIn {
	return lo.Map(msgs, func(msg *MessageIn, _ int) *MessageIn {
		return w.WrapMessageIn(subsystem, index, msg)
	})
}

func (w *MsgWrapper) DelegateInput(subsystem byte, index int, input Input) (GPA, []*MessageOut, error) {
	sub, err := w.subsystemFunc(subsystem, index)
	if err != nil {
		return nil, nil, err
	}
	return sub, w.WrapMessagesOut(subsystem, index, sub.Input(input)), nil
}

func (w *MsgWrapper) DelegateMessage(msg *TypedMessageIn[*WrappingMsg]) (GPA, []*MessageOut, error) {
	sub, err := w.subsystemFunc(msg.Payload.subsystem, msg.Payload.index)
	if err != nil {
		return nil, nil, err
	}
	subOut := sub.Message(NewMessageIn(msg.Sender, msg.Payload.wrapped))
	return sub, w.WrapMessagesOut(msg.Payload.subsystem, msg.Payload.index, subOut), nil
}

func (w *MsgWrapper) UnmarshalPayload(data []byte) (MessagePayload, error) {
	rawMsg, err := bcs.Unmarshal[rawWrappingMsg](data)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling wrapping msg: %w", err)
	}

	subGPA, err := w.subsystemFunc(rawMsg.Subsystem, rawMsg.Index)
	if err != nil {
		return nil, fmt.Errorf("retrieving subsystem GPA %v/%v: %w", rawMsg.Subsystem, rawMsg.Index, err)
	}

	wrapped, err := subGPA.UnmarshalPayload(rawMsg.WrappedMsgBytes)
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

// WrappingMsg is a message that contains another, and its routing info.
type WrappingMsg struct {
	msgType   MessageType
	subsystem byte
	index     int
	wrapped   MessagePayload
}

var _ MessagePayload = new(WrappingMsg)

func (msg *WrappingMsg) MsgType() MessageType {
	return msg.msgType
}

func (msg *WrappingMsg) Subsystem() byte {
	return msg.subsystem
}

func (msg *WrappingMsg) Index() int {
	return msg.index
}

func (msg *WrappingMsg) WrappedIn(sender NodeID) *MessageIn {
	return NewMessageIn(sender, msg.wrapped)
}

func (msg *WrappingMsg) WrappedOut(receipient NodeID) *MessageOut {
	return NewMessageOut(receipient, msg.wrapped)
}

func (msg *WrappingMsg) MarshalBCS(e *bcs.Encoder) error {
	wrappedMsgBytes, err := MarshalPayload(msg.wrapped)
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

func (msg *WrappingMsg) String() string {
	return fmt.Sprintf("WrappingMsg{subsystem=%v, index=%v, wrapped=%s}", msg.subsystem, msg.index, msg.wrapped)
}

type rawWrappingMsg struct {
	Subsystem       byte
	Index           int `bcs:"type=u16"`
	WrappedMsgBytes []byte
}
