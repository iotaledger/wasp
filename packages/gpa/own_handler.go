// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import "fmt"

// OwnHandler is a GPA instance handling own messages immediately.
type OwnHandler struct {
	me     NodeID
	target GPA
}

var _ GPA = &OwnHandler{}

func NewOwnHandler(me NodeID, target GPA) GPA {
	return &OwnHandler{me: me, target: target}
}

func (o *OwnHandler) Input(input Input) []Message {
	msgs := o.target.Input(input)
	myMsgs := []Message{}
	outMsgs := []Message{}
	for i, m := range msgs {
		if m.Recipient() == o.me {
			msgs[i].SetSender(o.me)
			myMsgs = append(myMsgs, msgs[i])
		} else {
			outMsgs = append(outMsgs, msgs[i])
		}
	}
	return o.handleMsgs(myMsgs, outMsgs)
}

func (o *OwnHandler) Message(msg Message) []Message {
	myMsgs := []Message{msg}
	outMsgs := []Message{}
	return o.handleMsgs(myMsgs, outMsgs)
}

func (o *OwnHandler) Output() Output {
	return o.target.Output()
}

func (o *OwnHandler) StatusString() string {
	return fmt.Sprintf("{OwnHandler, target=%s}", o.target.StatusString())
}

func (o *OwnHandler) UnmarshalMessage(data []byte) (Message, error) {
	return o.target.UnmarshalMessage(data)
}

func (o *OwnHandler) handleMsgs(myMsgs, outMsgs []Message) []Message {
	for len(myMsgs) > 0 {
		msgs := o.target.Message(myMsgs[0])
		myMsgs = myMsgs[1:]
		for i, m := range msgs {
			if m.Recipient() == o.me {
				msgs[i].SetSender(o.me)
				myMsgs = append(myMsgs, msgs[i])
			} else {
				outMsgs = append(outMsgs, msgs[i])
			}
		}
	}
	return outMsgs
}
