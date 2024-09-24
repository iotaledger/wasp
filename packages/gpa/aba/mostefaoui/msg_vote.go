// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mostefaoui

import (
	"github.com/iotaledger/wasp/packages/gpa"
)

type msgVoteType byte

const (
	BVAL msgVoteType = iota
	AUX
)

type msgVote struct {
	gpa.BasicMessage
	round    int         `bcs:"type=u16"`
	voteType msgVoteType `bcs:""`
	value    bool        `bcs:""`
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
