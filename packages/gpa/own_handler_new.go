// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import (
	"fmt"
	"slices"

	"github.com/iotaledger/hive.go/log"
)

// OwnHandlerNEW is a GPAnew instance handling own messages immediately.
//
// The idea is instead of checking if a message for myself in the actual
// protocols, one just send a message, and this handler passes it back
// as an ordinary message.
type OwnHandlerNEW struct {
	me     NodeID
	target GPA
}

var _ GPA = &OwnHandlerNEW{}

func NewOwnHandlerNEWWithOutPredicate(me NodeID, target GPA) GPA {
	return &OwnHandlerNEW{me: me, target: target}
}

func NewOwnHandlerNEW(me NodeID, target GPA) GPA {
	return NewOwnHandlerNEWWithOutPredicate(me, target)
}

func (o *OwnHandlerNEW) Input(input Input) []MessageOut {
	msgs := o.target.Input(input)
	return o.handleMsgs(msgs)
}

func (o *OwnHandlerNEW) Message(msg MessageIn[any]) []MessageOut {
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

func (o *OwnHandlerNEW) handleMsgs(msgs []MessageOut) []MessageOut {
	var outMsgs []MessageOut
	for len(msgs) > 0 {
		var msg MessageOut
		msg, msgs = msgs[0], msgs[1:]
		if msg.Recipient == o.me {
			// TODO: Review how can we avoid doing this marshal-unmarshal.
			// Currently it is needed, because message out payload is of type any,
			// while message in payload is of some specific type.

			b, err := o.MarshalPayload(msg.Payload)
			if err != nil {
				log.NewLogger().LogErrorf("failed to marshal own message payload: %#v: %v", msg.Payload, err)
				continue
			}

			payload, err := o.UnmarshalPayload(b)
			if err != nil {
				log.NewLogger().LogErrorf("failed to unmarshal own message payload: %v: %v", b, err)
				continue
			}

			msgs = slices.Concat(msgs, o.target.Message(NewMessageIn(o.me, payload)))
			continue
		}
		outMsgs = append(outMsgs, msg)
	}
	return outMsgs
}
