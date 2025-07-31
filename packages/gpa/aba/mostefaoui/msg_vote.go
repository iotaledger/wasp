// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mostefaoui

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

type msgVoteType byte

const (
	BVAL msgVoteType = iota
	AUX
)

type msgVote struct {
	gpa.BasicMessage
	round    int         `bcs:"export,type=u16"`
	voteType msgVoteType `bcs:"export"`
	value    bool        `bcs:"export"`
}

var _ gpa.Message = new(msgVote)

func multicastMsgVote(recipients []gpa.NodeID, round int, voteType msgVoteType, value bool) gpa.OutMessages {
	msgs := gpa.NoMessages()
	for _, recipient := range recipients {
		msgs.Add(&msgVote{
			BasicMessage: gpa.NewBasicMessage(recipient),
			round:        round,
			voteType:     voteType,
			value:        value,
		})
	}
	return msgs
}

func (msg *msgVote) MsgType() gpa.MessageType {
	return msgTypeVote
}
