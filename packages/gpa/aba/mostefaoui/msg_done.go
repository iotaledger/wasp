// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mostefaoui

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/gpa"
)

type msgDone struct {
	round int `bcs:"type=u16,export"`
}

var _ gpa.MessagePayload = new(msgDone)

func multicastMsgDone(recipients []gpa.NodeID, me gpa.NodeID, round int) []*gpa.MessageOut {
	var msgs []*gpa.MessageOut
	for _, recipient := range recipients {
		if recipient != me {
			msgs = append(msgs, gpa.NewMessageOut(recipient, &msgDone{
				round: round,
			}))
		}
	}
	return msgs
}

func (msg *msgDone) MsgType() gpa.MessageType {
	return msgTypeDone
}

func (msg *msgDone) String() string {
	return fmt.Sprintf("mostefaoui/Done(round=%d)", msg.round)
}
