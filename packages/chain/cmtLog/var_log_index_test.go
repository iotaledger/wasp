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
	vli := NewVarLogIndex(nodeIDs, n, f, initLI, log)
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

func randomAliasOutputWithID() *isc.AliasOutputWithID {
	outputID := testiotago.RandOutputID()
	aliasOutput := &iotago.AliasOutput{}
	return isc.NewAliasOutputWithID(aliasOutput, outputID)
}
