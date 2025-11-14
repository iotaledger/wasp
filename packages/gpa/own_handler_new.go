// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import (
	"fmt"
	"slices"
)

// OwnHandlerNEW is a GPAnew instance handling own messages immediately.
//
// The idea is instead of checking if a message for myself in the actual
// protocols, one just send a message, and this handler passes it back
// as an ordinary message.
type OwnHandlerNEW struct {
	me     NodeID
	target GPAnew
}

var _ GPAnew = &OwnHandlerNEW{}

func NewOwnHandlerNEWWithOutPredicate(me NodeID, target GPAnew) GPAnew {
	return &OwnHandlerNEW{me: me, target: target}
}

func NewOwnHandlerNEW(me NodeID, target GPAnew) GPAnew {
	return NewOwnHandlerNEWWithOutPredicate(me, target)
}

func (o *OwnHandlerNEW) Input(input Input) []PayloadOut {
	msgs := o.target.Input(input)
	return o.handleMsgs(msgs)
}

func (o *OwnHandlerNEW) Message(msg PayloadIn[any]) []PayloadOut {
	msgs := o.target.Message(msg)
	return o.handleMsgs(msgs)
}

func (o *OwnHandlerNEW) Output() Output {
	return o.target.Output()
}

func (o *OwnHandlerNEW) StatusString() string {
	return fmt.Sprintf("{OWN%s}", o.target.StatusString())
}

func (o *OwnHandlerNEW) MarshalPayload(payload any) ([]byte, error) {
	return o.target.MarshalPayload(payload)
}

func (o *OwnHandlerNEW) UnmarshalPayload(data []byte) (any, error) {
	return o.target.UnmarshalPayload(data)
}

func (o *OwnHandlerNEW) handleMsgs(msgs []PayloadOut) []PayloadOut {
	var outMsgs []PayloadOut
	for len(msgs) > 0 {
		var msg PayloadOut
		msg, msgs = msgs[0], msgs[1:]
		if msg.Recipient == o.me {
			msgs = slices.Concat(msgs, o.target.Message(NewPayloadIn(o.me, msg.Payload)))
			continue
		}
		outMsgs = append(outMsgs, msg)
	}
	return outMsgs
}
