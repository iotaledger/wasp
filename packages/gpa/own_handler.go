// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import (
	"fmt"
	"slices"
)

// OwnHandler is a GPA instance handling own messages immediately.
//
// The idea is instead of checking if a message for myself in the actual
// protocols, one just send a message, and this handler passes it back
// as an ordinary message.
type OwnHandler struct {
	me     NodeID
	target GPA
}

var _ GPA = &OwnHandler{}

func NewOwnHandlerWithOutPredicate(me NodeID, target GPA) GPA {
	return &OwnHandler{me: me, target: target}
}

func NewOwnHandler(me NodeID, target GPA) GPA {
	return NewOwnHandlerWithOutPredicate(me, target)
}

func (o *OwnHandler) Input(input Input) []MessageOut {
	msgs := o.target.Input(input)
	return o.handleMsgs(msgs)
}

func (o *OwnHandler) Message(msg MessageIn[any]) []MessageOut {
	msgs := o.target.Message(msg)
	return o.handleMsgs(msgs)
}

func (o *OwnHandler) Output() Output {
	return o.target.Output()
}

func (o *OwnHandler) StatusString() string {
	return fmt.Sprintf("{OWN%s}", o.target.StatusString())
}

func (o *OwnHandler) MarshalPayload(payload any) ([]byte, error) {
	return o.target.MarshalPayload(payload)
}

func (o *OwnHandler) UnmarshalPayload(data []byte) (any, error) {
	return o.target.UnmarshalPayload(data)
}

func (o *OwnHandler) handleMsgs(msgs []MessageOut) []MessageOut {
	var outMsgs []MessageOut
	for len(msgs) > 0 {
		var msg MessageOut
		msg, msgs = msgs[0], msgs[1:]
		if msg.Recipient == o.me {
			msgs = slices.Concat(msgs, o.target.Message(NewMessageIn(o.me, msg.Payload)))
			continue
		}
		outMsgs = append(outMsgs, msg)
	}
	return outMsgs
}
