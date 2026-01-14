// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acss

import (
	"testing"

	bcs "github.com/iotaledger/bcs-go"
)

func TestMsgVoteSerialization(t *testing.T) {
	{
		msg := &MsgVote{
			msgVoteOK,
		}

		bcs.TestCodecAndHash(t, msg, "93b889cd9f71")
	}
	{
		msg := &MsgVote{
			msgVoteREADY,
		}

		bcs.TestCodecAndHash(t, msg, "55a589aadf41")
	}
}
