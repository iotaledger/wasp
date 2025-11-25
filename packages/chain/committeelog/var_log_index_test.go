// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package committeelog_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/chain/committeelog"
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
	initLI := committeelog.NilLogIndex().Next()
	//
	vliOut := committeelog.NilLogIndex()
	vli := committeelog.NewVarLogIndex(nodeIDs, n, f, initLI, func(li committeelog.LogIndex) committeelog.OutMessages {
		vliOut = li
		return nil
	}, nil, log)
	//
	nextLI := initLI.Next()
	require.NotEqual(t, nextLI, vliOut)
	nextLIMsg := committeelog.NewMsgNextLogIndex(nextLI, committeelog.MsgNextLogIndexCauseStarted, false)
	for i := 0; i < n-f; i++ {
		vli.MsgNextLogIndexReceived(gpa.MessageIn[committeelog.MsgNextLogIndex]{
			Sender:  nodeIDs[i],
			Payload: *nextLIMsg,
		})
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
	initLI := committeelog.NilLogIndex().Next()
	//
	vliOut := committeelog.NilLogIndex()
	vli := committeelog.NewVarLogIndex(nodeIDs, n, f, initLI, func(li committeelog.LogIndex) committeelog.OutMessages {
		vliOut = li
		return nil
	}, nil, log)
	li15 := committeelog.LogIndex(15)
	li16 := committeelog.LogIndex(16)
	li18 := committeelog.LogIndex(18)
	require.Equal(t, committeelog.NilLogIndex(), vliOut)

	msgWithSender := func(sender gpa.NodeID, li committeelog.LogIndex) gpa.MessageIn[committeelog.MsgNextLogIndex] {
		msg := committeelog.NewMsgNextLogIndex(li, committeelog.MsgNextLogIndexCauseStarted, false)
		return gpa.MessageIn[committeelog.MsgNextLogIndex]{
			Sender:  sender,
			Payload: *msg,
		}
	}

	vli.MsgNextLogIndexReceived(msgWithSender(nodeIDs[0], li15))
	require.Equal(t, committeelog.NilLogIndex(), vliOut)

	vli.MsgNextLogIndexReceived(msgWithSender(nodeIDs[1], li18))
	require.Equal(t, li15, vliOut)

	vli.MsgNextLogIndexReceived(msgWithSender(nodeIDs[2], li16))
	require.Equal(t, li16, vliOut)

	vli.MsgNextLogIndexReceived(msgWithSender(nodeIDs[3], li15))
	require.Equal(t, li16, vliOut)
}
