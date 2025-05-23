// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtlog_test

import (
	"math/rand"
	"testing"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/chain/cmtlog"
	"github.com/iotaledger/wasp/packages/gpa"
)

func TestMsgNextLogIndexSerialization(t *testing.T) {
	{
		msg := &cmtlog.MsgNextLogIndex{
			gpa.BasicMessage{},
			cmtlog.LogIndex(rand.Int31()),
			cmtlog.MsgNextLogIndexCauseStarted,
			false,
		}

		bcs.TestCodec(t, msg)
	}
	msg := &cmtlog.MsgNextLogIndex{
		gpa.BasicMessage{},
		cmtlog.LogIndex(rand.Int31()),
		cmtlog.MsgNextLogIndexCauseStarted,
		true,
	}

	bcs.TestCodec(t, msg)
}
