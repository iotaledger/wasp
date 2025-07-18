// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acss

import (
	"testing"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/gpa"
)

func TestMsgVoteSerialization(t *testing.T) {
	{
		msg := &msgVote{
			gpa.BasicMessage{},
			msgVoteOK,
		}

		bcs.TestCodecAndHash(t, msg, "93b889cd9f71")
	}
	{
		msg := &msgVote{
			gpa.BasicMessage{},
			msgVoteREADY,
		}

		bcs.TestCodecAndHash(t, msg, "55a589aadf41")
	}
}
