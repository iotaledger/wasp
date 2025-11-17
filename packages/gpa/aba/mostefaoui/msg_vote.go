// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mostefaoui

import (
	"fmt"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/v2/packages/gpa"
)

type msgVoteType byte

const (
	BVAL msgVoteType = iota
	AUX
)

func (v msgVoteType) String() string {
	switch v {
	case BVAL:
		return "BVAL"
	case AUX:
		return "AUX"
	default:
		return "Unknown"
	}
}

type MsgVote struct {
	round    int         `bcs:"export,type=u16"`
	voteType msgVoteType `bcs:"export"`
	value    bool        `bcs:"export"`
}

func multicastMsgVote(recipients []gpa.NodeID, round int, voteType msgVoteType, value bool) []gpa.PayloadOut {
	return lo.Map(recipients, func(recipient gpa.NodeID, _ int) gpa.PayloadOut {
		return gpa.NewPayloadOut(recipient, MsgVote{
			round:    round,
			voteType: voteType,
			value:    value,
		})
	})
}

func (msg *MsgVote) String() string {
	return fmt.Sprintf("mostefaoui/Vote(round=%d, type=%s, value=%t)", msg.round, msg.voteType.String(), msg.value)
}
