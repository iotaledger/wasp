// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmt_log

import (
	"math/rand"
	"testing"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func TestMsgNextLogIndexSerialization(t *testing.T) {
	{
		msg := &MsgNextLogIndex{
			gpa.BasicMessage{},
			LogIndex(rand.Int31()),
			MsgNextLogIndexCauseRecover,
			false,
		}

		bcs.TestCodec(t, msg)
	}
	msg := &MsgNextLogIndex{
		gpa.BasicMessage{},
		LogIndex(rand.Int31()),
		MsgNextLogIndexCauseRecover,
		true,
	}

	bcs.TestCodec(t, msg)
}
