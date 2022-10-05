// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import "fmt"

// OwnHandler is a GPA instance handling own messages immediately.
//
// The idea is instead of checking if a message for myself in the actual
// protocols, one just send a message, and this handler passes it back
// as an ordinary message.
type OwnHandler struct {
	me           NodeID
	target       GPA
	outPredicate func(msg Message) bool
}

var _ GPA = &OwnHandler{}

func NewOwnHandlerWithOutPredicate(me NodeID, target GPA, outPredicate func(Message) bool) GPA {
	return &OwnHandler{me: me, target: target, outPredicate: outPredicate}
}

func NewOwnHandler(me NodeID, target GPA) GPA {
	return NewOwnHandlerWithOutPredicate(me, target, func(msg Message) bool { return false })
}

func (o *OwnHandler) Input(input Input) OutMessages {
	msgs := o.target.Input(input)
	outMsgs := NoMessages()
	return o.handleMsgs(msgs, outMsgs)
}

func (o *OwnHandler) Message(msg Message) OutMessages {
	msgs := o.target.Message(msg)
	outMsgs := NoMessages()
	return o.handleMsgs(msgs, outMsgs)
}

func (o *OwnHandler) Output() Output {
	return o.target.Output()
}

func (o *OwnHandler) StatusString() string {
	return fmt.Sprintf("{OWN/%v/%s}", o.me, o.target.StatusString())
}

func (o *OwnHandler) UnmarshalMessage(data []byte) (Message, error) {
	return o.target.UnmarshalMessage(data)
}

func (o *OwnHandler) handleMsgs(msgs, outMsgs OutMessages) OutMessages {
	if msgs == nil {
		return outMsgs
	}
	msgs.MustIterate(func(msg Message) {
		if msg.Recipient() == o.me && !o.outPredicate(msg) {
			msg.SetSender(o.me)
			msgs.AddAll(o.target.Message(msg))
		} else {
			outMsgs.Add(msg)
		}
	})
	return outMsgs
}
