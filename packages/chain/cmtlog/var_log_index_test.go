// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtlog_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/chain/cmtlog"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/testutil/testlogger"
)

func TestVarLogIndexV2Basic(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Shutdown()
	n := 4
	f := 1
	//
	nodeIDs := gpa.MakeTestNodeIDs(4)
	initLI := cmtlog.NilLogIndex().Next()
	//
	vliOut := cmtlog.NilLogIndex()
	vli := cmtlog.NewVarLogIndex(nodeIDs, n, f, initLI, func(li cmtlog.LogIndex) gpa.OutMessages {
		vliOut = li
		return nil
	}, nil, log)
	//
	nextLI := initLI.Next()
	require.NotEqual(t, nextLI, vliOut)
	nextLIMsg := cmtlog.NewMsgNextLogIndex(nodeIDs[0], nextLI, cmtlog.MsgNextLogIndexCauseStarted, false)
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
	initLI := cmtlog.NilLogIndex().Next()
	//
	vliOut := cmtlog.NilLogIndex()
	vli := cmtlog.NewVarLogIndex(nodeIDs, n, f, initLI, func(li cmtlog.LogIndex) gpa.OutMessages {
		vliOut = li
		return nil
	}, nil, log)
	li15 := cmtlog.LogIndex(15)
	li16 := cmtlog.LogIndex(16)
	li18 := cmtlog.LogIndex(18)
	require.Equal(t, cmtlog.NilLogIndex(), vliOut)

	msgWithSender := func(sender gpa.NodeID, li cmtlog.LogIndex) *cmtlog.MsgNextLogIndex {
		msg := cmtlog.NewMsgNextLogIndex(nodeIDs[0], li, cmtlog.MsgNextLogIndexCauseStarted, false)
		msg.SetSender(sender)
		return msg
	}

	vli.MsgNextLogIndexReceived(msgWithSender(nodeIDs[0], li15))
	require.Equal(t, cmtlog.NilLogIndex(), vliOut)

	vli.MsgNextLogIndexReceived(msgWithSender(nodeIDs[1], li18))
	require.Equal(t, li15, vliOut)

	vli.MsgNextLogIndexReceived(msgWithSender(nodeIDs[2], li16))
	require.Equal(t, li16, vliOut)

	vli.MsgNextLogIndexReceived(msgWithSender(nodeIDs[3], li15))
	require.Equal(t, li16, vliOut)
}
