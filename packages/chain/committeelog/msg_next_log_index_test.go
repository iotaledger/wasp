// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package committeelog_test

import (
	"math/rand"
	"testing"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/chain/committeelog"
)

func TestMsgNextLogIndexSerialization(t *testing.T) {
	{
		msg := &committeelog.MsgNextLogIndex{
			committeelog.LogIndex(rand.Int31()),
			committeelog.MsgNextLogIndexCauseStarted,
			false,
		}

		bcs.TestCodec(t, msg)
	}
	{
		msg := &committeelog.MsgNextLogIndex{
			committeelog.LogIndex(758493),
			committeelog.MsgNextLogIndexCauseStarted,
			false,
		}

		bcs.TestCodecAndHash(t, msg, "ad96fc92cd96")
	}
	{
		msg := &committeelog.MsgNextLogIndex{
			committeelog.LogIndex(rand.Int31()),
			committeelog.MsgNextLogIndexCauseStarted,
			true,
		}

		bcs.TestCodec(t, msg)
	}
	{
		msg := &committeelog.MsgNextLogIndex{
			committeelog.LogIndex(59329892),
			committeelog.MsgNextLogIndexCauseStarted,
			true,
		}

		bcs.TestCodecAndHash(t, msg, "c721637c3e91")
	}
}
