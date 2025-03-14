// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmt_log_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/chain/cmt_log"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

func TestVarLogIndexV2Basic(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Shutdown()
	n := 4
	f := 1
	//
	nodeIDs := gpa.MakeTestNodeIDs(4)
	initLI := cmt_log.NilLogIndex().Next()
	//
	vliOut := cmt_log.NilLogIndex()
	vli := cmt_log.NewVarLogIndex(nodeIDs, n, f, initLI, func(li cmt_log.LogIndex) gpa.OutMessages {
		vliOut = li
		return nil
	}, nil, log)
	//
	nextLI := initLI.Next()
	require.NotEqual(t, nextLI, vliOut)
	nextLIMsg := cmt_log.NewMsgNextLogIndex(nodeIDs[0], nextLI, cmt_log.MsgNextLogIndexCauseStarted, false)
	for i := 0; i < n-f; i++ {
		nextLIMsg.SetSender(nodeIDs[i])
		vli.MsgNextLogIndexReceived(nextLIMsg)
	}
	require.Equal(t, nextLI, vliOut)
}

func TestVarLogIndexV2Other(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Shutdown()
	n := 4
	f := 1
	//
	nodeIDs := gpa.MakeTestNodeIDs(4)
	initLI := cmt_log.NilLogIndex().Next()
	//
	vliOut := cmt_log.NilLogIndex()
	vli := cmt_log.NewVarLogIndex(nodeIDs, n, f, initLI, func(li cmt_log.LogIndex) gpa.OutMessages {
		vliOut = li
		return nil
	}, nil, log)
	li15 := cmt_log.LogIndex(15)
	li16 := cmt_log.LogIndex(16)
	li18 := cmt_log.LogIndex(18)
	require.Equal(t, cmt_log.NilLogIndex(), vliOut)

	msgWithSender := func(sender gpa.NodeID, li cmt_log.LogIndex) *cmt_log.MsgNextLogIndex {
		msg := cmt_log.NewMsgNextLogIndex(nodeIDs[0], li, cmt_log.MsgNextLogIndexCauseStarted, false)
		msg.SetSender(sender)
		return msg
	}

	vli.MsgNextLogIndexReceived(msgWithSender(nodeIDs[0], li15))
	require.Equal(t, cmt_log.NilLogIndex(), vliOut)

	vli.MsgNextLogIndexReceived(msgWithSender(nodeIDs[1], li18))
	require.Equal(t, li15, vliOut)

	vli.MsgNextLogIndexReceived(msgWithSender(nodeIDs[2], li16))
	require.Equal(t, li16, vliOut)

	vli.MsgNextLogIndexReceived(msgWithSender(nodeIDs[3], li15))
	require.Equal(t, li16, vliOut)
}
