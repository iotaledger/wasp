// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acss

import (
	"testing"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func TestMsgVoteSerialization(t *testing.T) {
	{
		msg := &msgVote{
			gpa.BasicMessage{},
			msgVoteOK,
		}

		bcs.TestCodec(t, msg)
	}
	{
		msg := &msgVote{
			gpa.BasicMessage{},
			msgVoteREADY,
		}

		bcs.TestCodec(t, msg)
	}
}
