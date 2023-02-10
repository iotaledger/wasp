// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtLog

import (
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil/testiotago"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

func TestVarLogIndex(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()
	n := 4
	f := 1
	//
	ao := randomAliasOutputWithID()
	nodeIDs := gpa.MakeTestNodeIDs(4)
	initLI := NilLogIndex().Next()
	//
	vli := NewVarLogIndex(nodeIDs, n, f, initLI, func(li LogIndex, ao *isc.AliasOutputWithID) {}, log)
	//
	nextLI := initLI.Next()
	vliLI, _ := vli.Value()
	require.NotEqual(t, nextLI, vliLI)
	nextLIMsg := newMsgNextLogIndex(nodeIDs[0], nextLI, ao)
	for i := 0; i < n-f; i++ {
		nextLIMsg.SetSender(nodeIDs[i])
		vli.MsgNextLogIndexReceived(nextLIMsg)
	}
	vliLI, _ = vli.Value()
	require.Equal(t, nextLI, vliLI)
}

func TestVarLogIndexV2(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()
	n := 4
	f := 1
	//
	ao := randomAliasOutputWithID()
	nodeIDs := gpa.MakeTestNodeIDs(4)
	initLI := NilLogIndex().Next()
	//
	vli := NewVarLogIndex(nodeIDs, n, f, initLI, func(li LogIndex, ao *isc.AliasOutputWithID) {}, log)
	vliValueLI := func() LogIndex {
		li, _ := vli.Value()
		return li
	}
	li15 := LogIndex(15)
	li16 := LogIndex(16)
	li18 := LogIndex(18)
	require.Equal(t, NilLogIndex(), vliValueLI())

	msgWithSender := func(sender gpa.NodeID, li LogIndex) *msgNextLogIndex {
		msg := newMsgNextLogIndex(nodeIDs[0], li, ao)
		msg.SetSender(sender)
		return msg
	}

	vli.MsgNextLogIndexReceived(msgWithSender(nodeIDs[0], li15))
	require.Equal(t, NilLogIndex(), vliValueLI())

	vli.MsgNextLogIndexReceived(msgWithSender(nodeIDs[1], li18))
	require.Equal(t, NilLogIndex(), vliValueLI())

	vli.MsgNextLogIndexReceived(msgWithSender(nodeIDs[2], li16))
	require.Equal(t, li15, vliValueLI())

	vli.MsgNextLogIndexReceived(msgWithSender(nodeIDs[3], li15))
	require.Equal(t, li15, vliValueLI())
}

func randomAliasOutputWithID() *isc.AliasOutputWithID {
	outputID := testiotago.RandOutputID()
	aliasOutput := &iotago.AliasOutput{}
	return isc.NewAliasOutputWithID(aliasOutput, outputID)
}
