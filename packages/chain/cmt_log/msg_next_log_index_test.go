// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmt_log

import (
	"math/rand"
	"testing"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestMsgNextLogIndexSerialization(t *testing.T) {
	{
		msg := &MsgNextLogIndex{
			gpa.BasicMessage{},
			LogIndex(rand.Int31()),
			isc.RandomAliasOutputWithID(),
			false,
		}

		rwutil.ReadWriteTest(t, msg, new(MsgNextLogIndex))
	}
	msg := &MsgNextLogIndex{
		gpa.BasicMessage{},
		LogIndex(rand.Int31()),
		isc.RandomAliasOutputWithID(),
		true,
	}

	rwutil.ReadWriteTest(t, msg, new(MsgNextLogIndex))
}
