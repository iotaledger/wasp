// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mostefaoui

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/gpa"
)

type MsgDone struct {
	round int `bcs:"type=u16,export"`
}

func multicastMsgDone(recipients []gpa.NodeID, me gpa.NodeID, round int) []gpa.MessageOut {
	var msgs []gpa.MessageOut
	for _, recipient := range recipients {
		if recipient != me {
			msgs = append(msgs, gpa.NewMessageOut(recipient, MsgDone{
				round: round,
			}))
		}
	}
	return msgs
}

func (msg *MsgDone) String() string {
	return fmt.Sprintf("mostefaoui/Done(round=%d)", msg.round)
}
