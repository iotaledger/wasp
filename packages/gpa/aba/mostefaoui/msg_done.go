// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mostefaoui

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

type msgDone struct {
	gpa.BasicMessage
	round int `bcs:"type=u16,export"`
}

var _ gpa.Message = new(msgDone)

func multicastMsgDone(recipients []gpa.NodeID, me gpa.NodeID, round int) gpa.OutMessages {
	msgs := gpa.NoMessages()
	for _, recipient := range recipients {
		if recipient != me {
			msgs.Add(&msgDone{
				BasicMessage: gpa.NewBasicMessage(recipient),
				round:        round,
			})
		}
	}
	return msgs
}

func (msg *msgDone) MsgType() gpa.MessageType {
	return msgTypeDone
}
