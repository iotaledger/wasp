// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acss

import (
	"testing"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestMsgVoteSerialization(t *testing.T) {
	{
		msg := &msgVote{
			gpa.BasicMessage{},
			msgVoteOK,
		}

		rwutil.ReadWriteTest(t, msg, new(msgVote))
	}
	{
		msg := &msgVote{
			gpa.BasicMessage{},
			msgVoteREADY,
		}

		rwutil.ReadWriteTest(t, msg, new(msgVote))
	}
}
