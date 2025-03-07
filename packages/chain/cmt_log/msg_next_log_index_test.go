// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmt_log_test

import (
	"math/rand"
	"testing"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/chain/cmt_log"
	"github.com/iotaledger/wasp/packages/gpa"
)

func TestMsgNextLogIndexSerialization(t *testing.T) {
	{
		msg := &cmt_log.MsgNextLogIndex{
			gpa.BasicMessage{},
			cmt_log.LogIndex(rand.Int31()),
			cmt_log.MsgNextLogIndexCauseStarted,
			false,
		}

		bcs.TestCodec(t, msg)
	}
	msg := &cmt_log.MsgNextLogIndex{
		gpa.BasicMessage{},
		cmt_log.LogIndex(rand.Int31()),
		cmt_log.MsgNextLogIndexCauseStarted,
		true,
	}

	bcs.TestCodec(t, msg)
}
