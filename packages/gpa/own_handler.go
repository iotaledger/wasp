// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import "fmt"

// OwnHandler is a GPA instance handling own messages immediately.
//
// The idea is instead of checking if a message for myself in the actual
// protocols, one just send a message, and this handler passes it back
// as an ordinary message.
type OwnHandler[Target GPA] struct {
	me           NodeID
	target       Target
	outPredicate func(msg Message) bool
}

var _ GPA = &OwnHandler[GPA]{}

func NewOwnHandlerWithOutPredicate[Target GPA](me NodeID, target Target, outPredicate func(Message) bool) *OwnHandler[Target] {
	return &OwnHandler[Target]{me: me, target: target, outPredicate: outPredicate}
}

func NewOwnHandler[Target GPA](me NodeID, target Target) *OwnHandler[Target] {
	return NewOwnHandlerWithOutPredicate(me, target, func(msg Message) bool { return false })
}

func (o *OwnHandler[_]) Input(input Input) OutMessages {
	msgs := o.target.Input(input)
	outMsgs := NoMessages()
	return o.handleMsgs(msgs, outMsgs)
}

func (o *OwnHandler[_]) Message(msg Message) OutMessages {
	msgs := o.target.Message(msg)
	outMsgs := NoMessages()
	return o.handleMsgs(msgs, outMsgs)
}

func (o *OwnHandler[_]) Output() Output {
	return o.target.Output()
}

func (o *OwnHandler[_]) StatusString() string {
	return fmt.Sprintf("{OWN%s}", o.target.StatusString())
}

func (o *OwnHandler[_]) UnmarshalMessage(data []byte) (Message, error) {
	return o.target.UnmarshalMessage(data)
}

func (o *OwnHandler[_]) handleMsgs(msgs, outMsgs OutMessages) OutMessages {
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
